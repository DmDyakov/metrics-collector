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

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
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
		_, err := p.db.ExecContext(ctx,
			`INSERT INTO counters (name, value, updated_dt) 
			VALUES ($1, $2, NOW())
			ON CONFLICT (name) DO UPDATE SET 
					value = EXCLUDED.value,
					updated_dt = NOW()`,
			m.ID, *m.Delta)
		return err

	case models.Gauge:
		if m.Value == nil {
			return nil
		}
		_, err := p.db.ExecContext(ctx,
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
	//TODO: Добавить реализацию
	return []models.Metrics{}, nil
}

func (p *PostgresStorage) saveMetricsBatch(ctx context.Context, metrics []models.Metrics) (*int, error) {
	count := 0
	tx, err := p.db.Begin()
	if err != nil {
		return &count, err
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
			count++

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
			count++
		}

	}
	tx.Commit()

	return &count, nil
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
