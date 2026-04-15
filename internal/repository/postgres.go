package repository

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
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
			return nil
		}
		_, err := p.db.ExecContext(ctx,
			`INSERT INTO counters (name, value) VALUES ($1, $2)
			 ON CONFLICT (name) DO UPDATE SET value = counters.value + EXCLUDED.value`,
			m.ID, *m.Delta)
		return err

	case models.Gauge:
		if m.Value == nil {
			return nil
		}
		_, err := p.db.ExecContext(ctx,
			`INSERT INTO gauges (name, value) VALUES ($1, $2)
			 ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value`,
			m.ID, *m.Value)
		return err
	}

	return nil
}

func (p *PostgresStorage) loadAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	//TODO: Добавить реализацию
	return []models.Metrics{}, nil
}

func (p *PostgresStorage) saveAllMetrics(ctx context.Context, metrics []models.Metrics) error {
	for _, m := range metrics {
		if err := p.saveMetric(ctx, &m); err != nil {
			return err
		}
	}

	return nil
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
