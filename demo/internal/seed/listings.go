package seed

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed seeds/demo_listings.sql
var demoListingsSeed string

// Listings populates an empty local demo database. It is deliberately
// separate from production migrations and is called only by the demo service.
func Listings(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin demo seed transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS demo_seed_state (
			key TEXT PRIMARY KEY,
			seeded_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`); err != nil {
		return fmt.Errorf("create demo seed state: %w", err)
	}

	var seeded bool
	err = tx.QueryRow(ctx, `
		INSERT INTO demo_seed_state (key)
		VALUES ('listings-v1')
		ON CONFLICT (key) DO NOTHING
		RETURNING true`).Scan(&seeded)
	if errors.Is(err, pgx.ErrNoRows) {
		return tx.Commit(ctx)
	}
	if err != nil {
		return fmt.Errorf("mark demo listings seed: %w", err)
	}

	if _, err := tx.Exec(ctx, demoListingsSeed); err != nil {
		return fmt.Errorf("seed demo listings: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit demo listings seed: %w", err)
	}
	return nil
}
