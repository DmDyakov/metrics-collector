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
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type PostgresStorage struct {
	db *sql.DB
}

func newPostgresStorage(databaseDSN string) (*PostgresStorage, error) {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ping := func() (struct{}, error) {
		return struct{}{}, db.PingContext(ctx)
	}
	_, err = doWithRetry(ctx, ping)
	if err != nil {
		return nil, err
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return &PostgresStorage{
		db: db,
	}, nil
}

func (p *PostgresStorage) saveMetric(ctx context.Context, m *models.Metrics) error {
	switch m.MType {
	case models.Counter:
		if m.Delta == nil {
			return errs.ErrMetricDeltaForCountRequired
		}

		saveCountMetric := func() (sql.Result, error) {
			return p.db.ExecContext(ctx,
				`INSERT INTO counters (name, value, updated_dt) 
				VALUES ($1, $2, NOW())
				ON CONFLICT (name) DO UPDATE SET 
					value = EXCLUDED.value,
					updated_dt = NOW()`,
				m.ID, *m.Delta)
		}

		_, err := doWithRetry(ctx, saveCountMetric)
		return err

	case models.Gauge:
		if m.Value == nil {
			return nil
		}

		saveGaugeMetric := func() (sql.Result, error) {
			return p.db.ExecContext(ctx,
				`INSERT INTO gauges (name, value, updated_dt) 
			VALUES ($1, $2, NOW())
				ON CONFLICT (name) DO UPDATE SET 
					value = EXCLUDED.value, 
					updated_dt = NOW()`,
				m.ID, *m.Value)
		}
		_, err := doWithRetry(ctx, saveGaugeMetric)
		return err
	}

	return nil
}

func (p *PostgresStorage) loadAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	queryAll := func() ([]models.Metrics, error) {
		var metrics []models.Metrics
		rows, err := p.db.QueryContext(ctx,
			`SELECT name, 'counter' as type, value::DOUBLE PRECISION as value FROM counters
			UNION ALL
			SELECT name, 'gauge' as type, value FROM gauges`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var m models.Metrics
			var value float64
			if err := rows.Scan(&m.ID, &m.MType, &value); err != nil {
				return nil, err
			}

			if m.MType == models.Counter {
				delta := int64(value)
				m.Delta = &delta
			} else if m.MType == models.Gauge {
				m.Value = &value
			}

			metrics = append(metrics, m)
		}

		return metrics, nil
	}

	return doWithRetry(ctx, queryAll)
}

func (p *PostgresStorage) saveMetricsBatch(ctx context.Context, metrics []models.Metrics) (*int, error) {
	saveMetrics := func() (*int, error) {
		savedMetricsCount := 0
		tx, err := p.db.Begin()
		if err != nil {
			return &savedMetricsCount, err
		}
		defer tx.Rollback()

		for _, m := range metrics {
			switch m.MType {
			case models.Counter:
				if m.Delta == nil {
					tx.Rollback()
					return nil, errs.ErrMetricDeltaForCountRequired
				}
				_, err := p.db.ExecContext(ctx,
					`INSERT INTO counters (name, value, updated_dt) 
				VALUES ($1, $2, NOW())
				ON CONFLICT (name) DO UPDATE SET 
					value = EXCLUDED.value,
					updated_dt = NOW()`,
					m.ID, *m.Delta)
				if err != nil {
					tx.Rollback()
					return nil, err
				}
				savedMetricsCount++

			case models.Gauge:
				if m.Value == nil {
					return nil, errs.ErrMetricValueForGaugeRequired
				}
				_, err := p.db.ExecContext(ctx,
					`INSERT INTO gauges (name, value, updated_dt) 
				VALUES ($1, $2, NOW())
			 	ON CONFLICT (name) DO UPDATE SET 
					value = EXCLUDED.value, 
					updated_dt = NOW()`,
					m.ID, *m.Value)
				if err != nil {
					tx.Rollback()
					return nil, err
				}
				savedMetricsCount++
			}

		}
		tx.Commit()
		return &savedMetricsCount, nil
	}

	return doWithRetry(ctx, saveMetrics)
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
		// Класс 08 - Connection Exception
		if pgerrcode.IsConnectionException(pgErr.Code) {
			return true
		}
	}
	return false
}

func doWithRetry[T any](ctx context.Context, fn func() (T, error)) (T, error) {
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
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(delay):
		}
	}
	return zero, nil
}
