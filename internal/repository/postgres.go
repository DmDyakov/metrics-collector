package repository

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"metrics-collector/internal/errs"
	models "metrics-collector/internal/model"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type DB struct {
	*sql.DB
	logger *zap.Logger
}
type PostgresStorage struct {
	db     DB
	logger *zap.Logger
}

func (db *DB) ExecContextWithRetry(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return doWithRetry(ctx, db.logger, func() (sql.Result, error) {
		return db.ExecContext(ctx, query, args...)
	})
}

func (db *DB) QueryContextWithRetry(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return doWithRetry(ctx, db.logger, func() (*sql.Rows, error) { return db.QueryContext(ctx, query, args...) })
}

//----------------------------------

func newPostgresStorage(databaseDSN string, logger *zap.Logger) (*PostgresStorage, error) {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return &PostgresStorage{
		db:     DB{db, logger},
		logger: logger,
	}, nil
}

func (p *PostgresStorage) saveMetric(ctx context.Context, m *models.Metrics) error {
	switch m.MType {
	case models.Counter:
		if m.Delta == nil {
			return errs.ErrMetricDeltaForCountRequired
		}

		_, err := p.db.ExecContextWithRetry(ctx,
			`INSERT INTO counters (name, value, updated_dt) 
				VALUES ($1, $2, NOW())
				ON CONFLICT (name) DO UPDATE SET 
					value = EXCLUDED.value,
					updated_dt = NOW()`,
			m.ID, *m.Delta)

		return err

	case models.Gauge:
		if m.Value == nil {
			return errs.ErrMetricValueForGaugeRequired
		}

		_, err := p.db.ExecContextWithRetry(ctx,
			`INSERT INTO gauges (name, value, updated_dt) 
			VALUES ($1, $2, NOW())
				ON CONFLICT (name) DO UPDATE SET 
					value = EXCLUDED.value, 
					updated_dt = NOW()`,
			m.ID, *m.Value)

		return err
	}

	return nil
}

func (p *PostgresStorage) loadAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	var metrics []models.Metrics
	rows, err := p.db.QueryContextWithRetry(ctx,
		`SELECT name, 'counter' as type, value::DOUBLE PRECISION as value FROM counters
			UNION ALL
			SELECT name, 'gauge' as type, value FROM gauges`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			p.logger.Error("rows close failed", zap.Error(err))
		}
	}()

	for rows.Next() {
		var m models.Metrics
		var value float64
		if err := rows.Scan(&m.ID, &m.MType, &value); err != nil {
			return nil, err
		}

		switch {
		case m.MType == models.Counter:
			delta := int64(value)
			m.Delta = &delta
		case m.MType == models.Gauge:
			m.Value = &value
		}

		metrics = append(metrics, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return metrics, nil
}

func (p *PostgresStorage) saveMetricsBatch(ctx context.Context, metrics []models.Metrics) (*int, error) {
	return doWithRetry(ctx, p.logger, func() (*int, error) {
		savedMetricsCount := 0
		tx, err := p.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		defer func() {
			if err != nil {
				if rbErr := tx.Rollback(); rbErr != nil {
					p.logger.Error("rollback failed", zap.Error(rbErr))
				}
			}
		}()

		for _, m := range metrics {
			switch m.MType {
			case models.Counter:
				if m.Delta == nil {
					return nil, errs.ErrMetricDeltaForCountRequired
				}
				_, err := tx.ExecContext(ctx,
					`INSERT INTO counters (name, value, updated_dt) 
				VALUES ($1, $2, NOW())
				ON CONFLICT (name) DO UPDATE SET 
					value = EXCLUDED.value,
					updated_dt = NOW()`,
					m.ID, *m.Delta)
				if err != nil {
					return nil, err
				}
				savedMetricsCount++

			case models.Gauge:
				if m.Value == nil {
					return nil, errs.ErrMetricValueForGaugeRequired
				}
				_, err := tx.ExecContext(ctx,
					`INSERT INTO gauges (name, value, updated_dt) 
				VALUES ($1, $2, NOW())
			 	ON CONFLICT (name) DO UPDATE SET 
					value = EXCLUDED.value, 
					updated_dt = NOW()`,
					m.ID, *m.Value)
				if err != nil {
					return nil, err
				}
				savedMetricsCount++
			}

		}

		err = tx.Commit()
		if err != nil {
			return nil, err
		}

		return &savedMetricsCount, err
	})

}

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

// ---- utils --------

func isRetriableDBError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgerrcode.IsConnectionException(pgErr.Code) {
			return true
		}

		switch pgErr.Code {
		case "40001", "40P01", "57P01", "53300", "55P03":
			return true
		}
	}
	return false
}

func doWithRetry[T any](ctx context.Context, logger *zap.Logger, fn func() (T, error)) (T, error) {
	delays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	const maxAttempts = 4

	var zero T
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		if !isRetriableDBError(err) || attempt == maxAttempts {
			return zero, err
		}

		delay := delays[attempt-1]
		logger.Warn("retrying db operation due to retriable error",
			zap.Int("attempt", attempt),
			zap.Duration("delay", delay),
			zap.Error(err),
		)

		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(delay):
		}
	}
	return zero, nil
}
