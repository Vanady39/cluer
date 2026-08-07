package controllers_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Vanady39/cluer/internal/controllers"
	"github.com/Vanady39/cluer/internal/domains"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubTourDomain struct {
	created *domains.Tour
	tours   []*domains.Tour
	err     error
}

func (s *stubTourDomain) Create(_ context.Context, t *domains.Tour) (*domains.Tour, error) {
	if s.err != nil {
		return nil, s.err
	}
	t.Id = uuid.Must(uuid.NewV7())
	s.created = t
	return t, nil
}

func (s *stubTourDomain) ListPublished(_ context.Context, _ string) ([]*domains.Tour, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.tours, nil
}

func tourRouter(domain domains.TourDomainInterface) *gin.Engine {
	controller := controllers.NewTourController(domain)
	return newTestRouter(func(r *gin.Engine) {
		r.POST("/tours", controller.Create)
		r.GET("/tours/published", controller.GetPublished)
	})
}

func TestTourController_Create(t *testing.T) {
	t.Run("valid tour returns 201 with Location and no body", func(t *testing.T) {
		domain := &stubTourDomain{}
		rec := doJSON(t, tourRouter(domain), http.MethodPost, "/tours", map[string]any{
			"title":       "Тур",
			"target_path": "/dashboard",
			"priority":    1,
		})

		require.Equal(t, http.StatusCreated, rec.Code)
		require.NotNil(t, domain.created)
		assert.Equal(t, "/v1/tours/"+domain.created.Id.String(), rec.Header().Get("Location"))
		assert.Empty(t, rec.Body.String())
	})

	t.Run("malformed JSON returns 400 in the standard envelope", func(t *testing.T) {
		rec := doRaw(tourRouter(&stubTourDomain{}), http.MethodPost, "/tours", "{not json")

		require.Equal(t, http.StatusBadRequest, rec.Code)
		payload := decodeHTTPError(t, rec)
		assert.Contains(t, payload.Error, "bind")
	})

	t.Run("domain validation error keeps its status and reason", func(t *testing.T) {
		domain := &stubTourDomain{err: &domains.LogicError{
			Err:   domains.ErrTitleRequired,
			Stage: "tour validation",
			Code:  http.StatusBadRequest,
		}}
		rec := doJSON(t, tourRouter(domain), http.MethodPost, "/tours", map[string]any{
			"target_path": "/dashboard",
		})

		require.Equal(t, http.StatusBadRequest, rec.Code)
		payload := decodeHTTPError(t, rec)
		assert.Contains(t, payload.Error, domains.ErrTitleRequired.Error())
	})

	t.Run("repository error surfaces as 404 for a missing record", func(t *testing.T) {
		domain := &stubTourDomain{err: &domains.RepositoryError{
			Err:       domains.ErrTourNotFound,
			EntityRef: "abc",
			Operation: domains.Retrieve,
			Reason:    domains.NoRecord,
		}}
		rec := doJSON(t, tourRouter(domain), http.MethodPost, "/tours", map[string]any{"title": "Тур"})

		require.Equal(t, http.StatusNotFound, rec.Code)
		decodeHTTPError(t, rec)
	})
}

func TestTourController_GetPublished(t *testing.T) {
	t.Run("returns tours for the requested path", func(t *testing.T) {
		domain := &stubTourDomain{tours: []*domains.Tour{
			{Id: uuid.Must(uuid.NewV7()), Title: "Тур", TargetPath: "/dashboard", Status: domains.TourPublished},
		}}
		rec := doJSON(t, tourRouter(domain), http.MethodGet, "/tours/published?path=/dashboard", nil)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Тур")
	})

	t.Run("missing path query returns 400", func(t *testing.T) {
		rec := doJSON(t, tourRouter(&stubTourDomain{}), http.MethodGet, "/tours/published", nil)

		require.Equal(t, http.StatusBadRequest, rec.Code)
		payload := decodeHTTPError(t, rec)
		assert.Contains(t, payload.Error, "query")
	})

	t.Run("an unrecognized error type falls back to 500", func(t *testing.T) {
		domain := &stubTourDomain{err: assert.AnError}
		rec := doJSON(t, tourRouter(domain), http.MethodGet, "/tours/published?path=/x", nil)

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		decodeHTTPError(t, rec)
	})
}
