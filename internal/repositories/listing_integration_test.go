package repositories_test

import (
	"context"
	"os"
	"testing"

	"github.com/Vanady39/cluer/internal/domains"
	"github.com/Vanady39/cluer/internal/repositories"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListingRepositorySearch(t *testing.T) {
	dsn := os.Getenv("CLUER_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set CLUER_TEST_DATABASE_URL to run PostgreSQL integration tests")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	repo := repositories.NewListingRepository(pool)

	t.Run("empty q returns all listings", func(t *testing.T) {
		listings, err := repo.GetListings(ctx, domains.ListingFilter{Limit: 20})
		require.NoError(t, err)
		assert.NotEmpty(t, listings)
	})

	t.Run("title search is case-insensitive", func(t *testing.T) {
		listings, err := repo.GetListings(ctx, domains.ListingFilter{Query: "IPHONE", Limit: 20})
		require.NoError(t, err)
		require.Len(t, listings, 1)
		assert.Equal(t, "iPhone 15 Pro", listings[0].Title)
	})

	t.Run("description search finds a matching listing", func(t *testing.T) {
		listings, err := repo.GetListings(ctx, domains.ListingFilter{Query: "ЗАВОДСКОЙ УПАКОВКЕ", Limit: 20})
		require.NoError(t, err)
		require.Len(t, listings, 1)
		assert.Equal(t, "Samsung Galaxy S24", listings[0].Title)
	})

	t.Run("no match returns an empty array", func(t *testing.T) {
		listings, err := repo.GetListings(ctx, domains.ListingFilter{Query: "несуществующее объявление", Limit: 20})
		require.NoError(t, err)
		assert.Empty(t, listings)
		assert.NotNil(t, listings)
	})
}
