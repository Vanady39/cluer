package controllers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/Vanady39/cluer/internal/controllers"
	"github.com/Vanady39/cluer/internal/domains"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubListingDomain struct {
	page   domains.ListingPage
	err    error
	filter domains.ListingFilter
}

func (s *stubListingDomain) GetListings(_ context.Context, filter domains.ListingFilter) (domains.ListingPage, error) {
	s.filter = filter
	if s.err != nil {
		return domains.ListingPage{}, s.err
	}
	return s.page, nil
}

func listingRouter(domain domains.ListingDomainInterface) *gin.Engine {
	controller := controllers.NewListingController(domain)
	return newTestRouter(func(r *gin.Engine) {
		r.GET("/listings", controller.GetListings)
	})
}

func TestListingController_GetListings(t *testing.T) {
	t.Run("returns listings wrapped in a data envelope", func(t *testing.T) {
		domain := &stubListingDomain{page: domains.ListingPage{
			Listings: []domains.Listing{{ID: 1, Title: "iPhone", Description: "как новый", Price: 95000, ImageURL: "https://x/i.png"}},
		}}
		rec := doJSON(t, listingRouter(domain), http.MethodGet, "/listings", nil)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t,
			`{"data":[{"id":1,"title":"iPhone","description":"как новый","price":95000,"imageUrl":"https://x/i.png"}]}`,
			rec.Body.String(),
		)
		assert.Equal(t, domains.ListingFilter{Limit: domains.DefaultListingsLimit}, domain.filter)
	})

	t.Run("trims q and passes pagination filter to the domain", func(t *testing.T) {
		domain := &stubListingDomain{}
		rec := doJSON(t, listingRouter(domain), http.MethodGet, "/listings?q=+iPhone+&limit=5&offset=2", nil)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, domains.ListingFilter{Query: "iPhone", Limit: 5, Offset: 2}, domain.filter)
	})

	t.Run("empty result still returns a data array", func(t *testing.T) {
		rec := doJSON(t, listingRouter(&stubListingDomain{page: domains.ListingPage{Listings: []domains.Listing{}}}), http.MethodGet, "/listings", nil)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"data":[]}`, rec.Body.String())
	})

	t.Run("domain failure returns 500 in the standard envelope", func(t *testing.T) {
		rec := doJSON(t, listingRouter(&stubListingDomain{err: assert.AnError}), http.MethodGet, "/listings", nil)

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		decodeHTTPError(t, rec)
	})

	for _, path := range []string{
		"/listings?limit=",
		"/listings?offset=",
		"/listings?limit=0",
		"/listings?limit=51",
		"/listings?limit=invalid",
		"/listings?offset=-1",
		"/listings?offset=100001",
		"/listings?offset=invalid",
		"/listings?q=" + strings.Repeat("q", domains.MaxListingsQueryRunes+1),
	} {
		t.Run("pagination validation: "+path, func(t *testing.T) {
			rec := doJSON(t, listingRouter(&stubListingDomain{}), http.MethodGet, path, nil)

			if path == "/listings?limit=" || path == "/listings?offset=" {
				require.Equal(t, http.StatusOK, rec.Code)
				return
			}
			require.Equal(t, http.StatusBadRequest, rec.Code)
			payload := decodeHTTPError(t, rec)
			assert.Equal(t, "failed to bind query", payload.Error)
			assert.Contains(t, payload.Message, "Failed to parse: ")
		})
	}
}

// TestListingController_GetListingsDocContract проверяет контракт из
// listings-api.md (части 1 и 2), а не текущее поведение кода.
func TestListingController_GetListingsDocContract(t *testing.T) {
	sample := domains.Listing{
		ID:          3,
		Title:       "Samsung Galaxy S24",
		Description: "Новый телефон в заводской упаковке",
		Price:       72000,
		ImageURL:    "https://example.com/samsung.png",
	}

	t.Run("успешный ответ состоит ровно из поля data", func(t *testing.T) {
		domain := &stubListingDomain{page: domains.ListingPage{Listings: []domains.Listing{sample}}}
		rec := doJSON(t, listingRouter(domain), http.MethodGet, "/listings", nil)

		require.Equal(t, http.StatusOK, rec.Code)

		var envelope map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))

		keys := make([]string, 0, len(envelope))
		for key := range envelope {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		assert.Equal(t, []string{"data"}, keys,
			"по доке конверт состоит только из data (total — открытый вопрос №1): %s", rec.Body.String())
	})

	t.Run("price отдаётся целым числом рублей", func(t *testing.T) {
		domain := &stubListingDomain{page: domains.ListingPage{Listings: []domains.Listing{sample}}}
		rec := doJSON(t, listingRouter(domain), http.MethodGet, "/listings", nil)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"price":72000`)
	})

	t.Run("порядок объявлений в ответе повторяет id DESC из домена", func(t *testing.T) {
		domain := &stubListingDomain{page: domains.ListingPage{Listings: []domains.Listing{
			{ID: 3, Title: "Samsung Galaxy S24"},
			{ID: 2, Title: "Игровое кресло"},
			{ID: 1, Title: "iPhone 15 Pro"},
		}}}
		rec := doJSON(t, listingRouter(domain), http.MethodGet, "/listings", nil)

		require.Equal(t, http.StatusOK, rec.Code)

		var payload struct {
			Data []struct {
				ID int64 `json:"id"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
		require.Len(t, payload.Data, 3)
		assert.Equal(t, []int64{3, 2, 1}, []int64{payload.Data[0].ID, payload.Data[1].ID, payload.Data[2].ID})
	})

	for name, target := range map[string]string{
		"пустое q равносильно отсутствию параметра":      "/listings?q=",
		"q из одних пробелов равносильно пустому q":      "/listings?q=+++",
		"пустые limit и offset подставляют дефолты":      "/listings?limit=&offset=",
		"отсутствующие параметры дают limit=20 offset=0": "/listings",
	} {
		t.Run(name, func(t *testing.T) {
			domain := &stubListingDomain{}
			rec := doJSON(t, listingRouter(domain), http.MethodGet, target, nil)

			require.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t,
				domains.ListingFilter{Query: "", Limit: domains.DefaultListingsLimit, Offset: 0},
				domain.filter,
			)
		})
	}

	t.Run("границы limit 1 и 50 принимаются", func(t *testing.T) {
		for _, limit := range []int{1, domains.MaxListingsLimit} {
			domain := &stubListingDomain{}
			rec := doJSON(t, listingRouter(domain), http.MethodGet, "/listings?limit="+strconv.Itoa(limit), nil)

			require.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, limit, domain.filter.Limit)
		}
	})

	t.Run("offset=0 принимается", func(t *testing.T) {
		domain := &stubListingDomain{}
		rec := doJSON(t, listingRouter(domain), http.MethodGet, "/listings?offset=0", nil)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, 0, domain.filter.Offset)
	})

	t.Run("q со спецсимволами не является ошибкой и уходит в домен как есть", func(t *testing.T) {
		domain := &stubListingDomain{}
		rec := doJSON(t, listingRouter(domain), http.MethodGet, "/listings?q=100%25_match", nil)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "100%_match", domain.filter.Query)
	})

	// Следующие два гарда — намеренное решение владельца сервиса: верхняя граница
	// offset защищает от глубокого скана, ограничение длины q — от гигантских
	// подстрок в LIKE. В listings-api.md они пока не описаны, поэтому здесь
	// зафиксировано фактическое поведение кода, а не текст контракта.
	t.Run("гард сверх контракта: offset больше 100000 отклоняется", func(t *testing.T) {
		rec := doJSON(t, listingRouter(&stubListingDomain{}), http.MethodGet, "/listings?offset=100001", nil)

		require.Equal(t, http.StatusBadRequest, rec.Code, "гард на глубокий скан: %s", rec.Body.String())
		payload := decodeHTTPError(t, rec)
		assert.Equal(t, "failed to bind query", payload.Error)
		assert.Equal(t, "Failed to parse: offset must be an integer between 0 and 100000", payload.Message)
	})

	t.Run("граница offset=100000 остаётся валидной", func(t *testing.T) {
		domain := &stubListingDomain{}
		rec := doJSON(t, listingRouter(domain), http.MethodGet, "/listings?offset="+strconv.Itoa(domains.MaxListingsOffset), nil)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, domains.MaxListingsOffset, domain.filter.Offset)
	})

	t.Run("гард сверх контракта: q длиннее 200 рун отклоняется", func(t *testing.T) {
		long := strings.Repeat("ф", domains.MaxListingsQueryRunes+1)
		rec := doJSON(t, listingRouter(&stubListingDomain{}), http.MethodGet, "/listings?q="+url.QueryEscape(long), nil)

		require.Equal(t, http.StatusBadRequest, rec.Code, "гард на длину q: %s", rec.Body.String())
		payload := decodeHTTPError(t, rec)
		assert.Equal(t, "failed to bind query", payload.Error)
		assert.Equal(t, "Failed to parse: q must not exceed 200 characters", payload.Message)
	})

	t.Run("граница q ровно в 200 рун остаётся валидной", func(t *testing.T) {
		domain := &stubListingDomain{}
		exact := strings.Repeat("ф", domains.MaxListingsQueryRunes)
		rec := doJSON(t, listingRouter(domain), http.MethodGet, "/listings?q="+url.QueryEscape(exact), nil)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, exact, domain.filter.Query)
	})

	t.Run("тело 400 совпадает с примером из доки", func(t *testing.T) {
		rec := doJSON(t, listingRouter(&stubListingDomain{}), http.MethodGet, "/listings?limit=abc", nil)

		require.Equal(t, http.StatusBadRequest, rec.Code)
		payload := decodeHTTPError(t, rec)
		assert.Equal(t, "failed to bind query", payload.Error)
		assert.NotEmpty(t, payload.Message)
	})

	t.Run("тело 500 при ошибке базы совпадает с примером из доки", func(t *testing.T) {
		repoErr := &domains.RepositoryError{
			Err:       assert.AnError,
			EntityRef: "listings",
			Operation: domains.Retrieve,
			Reason:    domains.QueryError,
		}
		rec := doJSON(t, listingRouter(&stubListingDomain{err: repoErr}), http.MethodGet, "/listings", nil)

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		payload := decodeHTTPError(t, rec)
		assert.Equal(t, "failed to retrieve entity listings: query", payload.Error)
		assert.Equal(t, "Unable to retrieve the requested entity", payload.Message)
	})
}
