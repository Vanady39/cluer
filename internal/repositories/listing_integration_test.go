package repositories

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/Vanady39/cluer/internal/domains"
	"github.com/google/uuid"
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

	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = tx.Rollback(ctx) })

	token := uuid.NewString()
	fixtures := []struct {
		title       string
		description string
	}{
		{"listing-search-" + token + " iPhone", "ordinary listing"},
		{"listing-search-" + token + " Samsung", "factory package " + token},
		{"listing-search-" + token + " literal 100%_match", "special characters"},
		{"listing-search-" + token + " literal 100xxmatch", "must not match literal wildcard search"},
	}

	for _, fixture := range fixtures {
		_, err := tx.Exec(ctx, `
			INSERT INTO listings (title, description, price, image_url)
			VALUES ($1, $2, 1, '')`, fixture.title, fixture.description)
		require.NoError(t, err)
	}

	repo := newListingRepository(tx)

	t.Run("empty q returns listings", func(t *testing.T) {
		page, err := repo.GetListings(ctx, domains.ListingFilter{Limit: 20})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, page.Total, int64(len(fixtures)))
		assert.NotEmpty(t, page.Listings)
	})

	t.Run("title search is case-insensitive", func(t *testing.T) {
		page, err := repo.GetListings(ctx, domains.ListingFilter{
			Query: strings.ToUpper("listing-search-" + token + " iphone"),
			Limit: 20,
		})
		require.NoError(t, err)
		require.Len(t, page.Listings, 1)
		assert.Equal(t, fixtures[0].title, page.Listings[0].Title)
		assert.EqualValues(t, 1, page.Total)
	})

	t.Run("description search finds a matching listing", func(t *testing.T) {
		page, err := repo.GetListings(ctx, domains.ListingFilter{Query: "FACTORY PACKAGE " + token, Limit: 20})
		require.NoError(t, err)
		require.Len(t, page.Listings, 1)
		assert.Equal(t, fixtures[1].title, page.Listings[0].Title)
	})

	t.Run("wildcards in q are literal", func(t *testing.T) {
		page, err := repo.GetListings(ctx, domains.ListingFilter{Query: "100%_match", Limit: 20})
		require.NoError(t, err)
		require.Len(t, page.Listings, 1)
		assert.Equal(t, fixtures[2].title, page.Listings[0].Title)
	})

	t.Run("no match returns an empty array", func(t *testing.T) {
		page, err := repo.GetListings(ctx, domains.ListingFilter{Query: "missing-" + token, Limit: 20})
		require.NoError(t, err)
		assert.Empty(t, page.Listings)
		assert.NotNil(t, page.Listings)
		assert.Zero(t, page.Total)
	})

	t.Run("total is returned for a partial page", func(t *testing.T) {
		page, err := repo.GetListings(ctx, domains.ListingFilter{Query: "listing-search-" + token, Limit: 2, Offset: 2})
		require.NoError(t, err)
		assert.Len(t, page.Listings, 2)
		assert.EqualValues(t, len(fixtures), page.Total)
	})
}
