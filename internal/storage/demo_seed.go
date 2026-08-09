package storage

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed seeds/demo_listings.sql
var demoListingsSeed string

// SeedDemoListings populates an empty local demo database. It is deliberately
// separate from production migrations and is called only by the demo service.
func SeedDemoListings(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, demoListingsSeed); err != nil {
		return fmt.Errorf("seed demo listings: %w", err)
	}
	return nil
}
