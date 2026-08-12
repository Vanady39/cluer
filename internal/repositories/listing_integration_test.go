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
		assert.GreaterOrEqual(t, len(page.Listings), len(fixtures))
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
	})

	t.Run("a partial page returns the remaining listings", func(t *testing.T) {
		first, err := repo.GetListings(ctx, domains.ListingFilter{Query: "listing-search-" + token, Limit: 2, Offset: 0})
		require.NoError(t, err)
		require.Len(t, first.Listings, 2)

		page, err := repo.GetListings(ctx, domains.ListingFilter{Query: "listing-search-" + token, Limit: 2, Offset: 2})
		require.NoError(t, err)
		assert.Len(t, page.Listings, 2)
		assert.Greater(t, first.Listings[1].ID, page.Listings[0].ID, "страницы не пересекаются и идут по id DESC")
	})
}

// TestListingRepositoryDocContract проверяет семантику поиска и пагинации так,
// как она описана в listings-api.md (части 1 и 2).
func TestListingRepositoryDocContract(t *testing.T) {
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
	prefix := "doc-contract-" + token
	fixtures := []struct {
		title       string
		description string
		price       int64
	}{
		{prefix + " iPhone 15 Pro", "Телефон в отличном состоянии, без царапин", 95000},
		{prefix + " Игровое кресло", "Использовалось полгода, удобная поддержка спины", 15000},
		{prefix + " Samsung Galaxy S24", "Новый телефон в заводской упаковке", 72000},
		{prefix + " подчёркивание a_b", "литеральный underscore", 100},
		{prefix + " подчёркивание axb", "не должен находиться по запросу a_b", 100},
		{prefix + " процент 50%", "литеральный percent", 100},
	}

	ids := make([]int64, 0, len(fixtures))
	for _, fixture := range fixtures {
		var id int64
		require.NoError(t, tx.QueryRow(ctx, `
			INSERT INTO listings (title, description, price, image_url)
			VALUES ($1, $2, $3, '')
			RETURNING id`, fixture.title, fixture.description, fixture.price).Scan(&id))
		ids = append(ids, id)
	}

	repo := newListingRepository(tx)

	titles := func(page domains.ListingPage) []string {
		out := make([]string, 0, len(page.Listings))
		for _, listing := range page.Listings {
			out = append(out, listing.Title)
		}
		return out
	}

	t.Run("выдача отсортирована по id DESC", func(t *testing.T) {
		page, err := repo.GetListings(ctx, domains.ListingFilter{Query: prefix, Limit: domains.MaxListingsLimit})
		require.NoError(t, err)
		require.Len(t, page.Listings, len(fixtures))

		for i := 1; i < len(page.Listings); i++ {
			assert.Greater(t, page.Listings[i-1].ID, page.Listings[i].ID,
				"ожидается id DESC, получено %v", page.Listings)
		}
		assert.Equal(t, ids[len(ids)-1], page.Listings[0].ID, "новые объявления идут первыми")
	})

	t.Run("кириллица ищется без учёта регистра", func(t *testing.T) {
		page, err := repo.GetListings(ctx, domains.ListingFilter{Query: "ТЕЛЕФОН " + prefix, Limit: 20})
		require.NoError(t, err)
		assert.Empty(t, page.Listings, "поиск идёт по подстроке, а не по словам")

		page, err = repo.GetListings(ctx, domains.ListingFilter{Query: prefix + " IGROVOE", Limit: 20})
		require.NoError(t, err)
		assert.Empty(t, page.Listings)

		page, err = repo.GetListings(ctx, domains.ListingFilter{Query: "ТЕЛЕФОН", Limit: domains.MaxListingsLimit})
		require.NoError(t, err)
		found := 0
		for _, listing := range page.Listings {
			if strings.HasPrefix(listing.Title, prefix) {
				found++
			}
		}
		assert.Equal(t, 2, found, "ТЕЛЕФОН должен находить «телефон» в описаниях: %v", titles(page))
	})

	t.Run("латиница в нижнем регистре находит запись из title", func(t *testing.T) {
		page, err := repo.GetListings(ctx, domains.ListingFilter{Query: strings.ToLower(prefix + " SAMSUNG"), Limit: 20})
		require.NoError(t, err)
		require.Len(t, page.Listings, 1)
		assert.Equal(t, fixtures[2].title, page.Listings[0].Title)
	})

	t.Run("подчёркивание в q ищется литерально", func(t *testing.T) {
		page, err := repo.GetListings(ctx, domains.ListingFilter{Query: prefix + " подчёркивание a_b", Limit: 20})
		require.NoError(t, err)
		require.Len(t, page.Listings, 1)
		assert.Equal(t, fixtures[3].title, page.Listings[0].Title)
	})

	t.Run("процент в q ищется литерально", func(t *testing.T) {
		page, err := repo.GetListings(ctx, domains.ListingFilter{Query: prefix + " процент 50%", Limit: 20})
		require.NoError(t, err)
		require.Len(t, page.Listings, 1)
		assert.Equal(t, fixtures[5].title, page.Listings[0].Title)
	})

	t.Run("поиск идёт по подстроке в середине текста", func(t *testing.T) {
		page, err := repo.GetListings(ctx, domains.ListingFilter{Query: "полгода, удобная", Limit: domains.MaxListingsLimit})
		require.NoError(t, err)
		assert.Contains(t, titles(page), fixtures[1].title)
	})

	t.Run("страницы limit/offset не пересекаются и идут по id DESC", func(t *testing.T) {
		first, err := repo.GetListings(ctx, domains.ListingFilter{Query: prefix, Limit: 2, Offset: 0})
		require.NoError(t, err)
		require.Len(t, first.Listings, 2)

		second, err := repo.GetListings(ctx, domains.ListingFilter{Query: prefix, Limit: 2, Offset: 2})
		require.NoError(t, err)
		require.Len(t, second.Listings, 2)

		assert.Greater(t, first.Listings[1].ID, second.Listings[0].ID)

		last, err := repo.GetListings(ctx, domains.ListingFilter{Query: prefix, Limit: 2, Offset: 4})
		require.NoError(t, err)
		assert.Len(t, last.Listings, 2)

		beyond, err := repo.GetListings(ctx, domains.ListingFilter{Query: prefix, Limit: 2, Offset: len(fixtures)})
		require.NoError(t, err)
		assert.Empty(t, beyond.Listings, "последняя страница короче limit — признак конца выдачи")
		assert.NotNil(t, beyond.Listings)
	})

	t.Run("price возвращается целым числом рублей", func(t *testing.T) {
		page, err := repo.GetListings(ctx, domains.ListingFilter{Query: prefix + " iPhone", Limit: 20})
		require.NoError(t, err)
		require.Len(t, page.Listings, 1)
		assert.EqualValues(t, 95000, page.Listings[0].Price)
	})
}
