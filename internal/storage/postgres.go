// Package storage owns the database connection and schema lifecycle.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Vanady39/cluer/migrations"
	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx/v5" database/sql driver
	"github.com/rs/zerolog"
)

// Connect opens a pooled connection and blocks until the database answers a
// ping or the deadline passes. Postgres in docker-compose accepts TCP a second
// or two before it is ready to serve queries, so a single attempt at startup is
// a coin flip; retrying here is cheaper than a restart loop on the container.
func Connect(ctx context.Context, dsn string, logger *zerolog.Logger) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	cfg.MaxConns = 10
	cfg.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	deadline := time.Now().Add(30 * time.Second)
	for attempt := 1; ; attempt++ {
		err = pool.Ping(ctx)
		if err == nil {
			logger.Info().Int("attempt", attempt).Msg("Database is up")
			return pool, nil
		}
		if time.Now().After(deadline) || ctx.Err() != nil {
			pool.Close()
			return nil, fmt.Errorf("database unreachable after %d attempts: %w", attempt, err)
		}
		logger.Debug().Err(err).Int("attempt", attempt).Msg("Database not ready, retrying")

		select {
		case <-ctx.Done():
			pool.Close()
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

// Migrate applies every pending migration embedded in the binary.
func Migrate(dsn string, logger *zerolog.Logger) error {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}

	db, err := sql.Open("pgx/v5", dsn)
	if err != nil {
		return fmt.Errorf("open migration connection: %w", err)
	}
	defer db.Close()

	driver, err := migratepgx.WithInstance(db, &migratepgx.Config{})
	if err != nil {
		return fmt.Errorf("build migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "pgx5", driver)
	if err != nil {
		return fmt.Errorf("build migrator: %w", err)
	}

	// ErrNoChange means the schema is already current, which is the normal case
	// on every restart after the first.
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("read schema version: %w", err)
	}
	logger.Info().Uint("version", version).Bool("dirty", dirty).Msg("Schema is up to date")

	return nil
}
