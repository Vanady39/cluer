package controllers_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Vanady39/cluer/internal/controllers"
	"github.com/Vanady39/cluer/internal/domains"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubListingDomain struct {
	listings []domains.Listing
	err      error
}

func (s *stubListingDomain) GetListings(_ context.Context) ([]domains.Listing, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.listings, nil
}

func listingRouter(domain domains.ListingDomainInterface) *gin.Engine {
	controller := controllers.NewListingController(domain)
	return newTestRouter(func(r *gin.Engine) {
		r.GET("/listings", controller.GetListings)
	})
}

func TestListingController_GetListings(t *testing.T) {
	t.Run("returns listings wrapped in a data envelope", func(t *testing.T) {
		domain := &stubListingDomain{listings: []domains.Listing{
			{ID: 1, Title: "iPhone", Description: "как новый", Price: 95000, ImageURL: "https://x/i.png"},
		}}
		rec := doJSON(t, listingRouter(domain), http.MethodGet, "/listings", nil)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t,
			`{"data":[{"id":1,"title":"iPhone","description":"как новый","price":95000,"imageUrl":"https://x/i.png"}]}`,
			rec.Body.String(),
		)
	})

	t.Run("empty result still returns a data array", func(t *testing.T) {
		rec := doJSON(t, listingRouter(&stubListingDomain{}), http.MethodGet, "/listings", nil)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"data":[]}`, rec.Body.String())
	})

	t.Run("domain failure returns 500 in the standard envelope", func(t *testing.T) {
		rec := doJSON(t, listingRouter(&stubListingDomain{err: assert.AnError}), http.MethodGet, "/listings", nil)

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		decodeHTTPError(t, rec)
	})
}
