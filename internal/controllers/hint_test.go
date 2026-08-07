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

type stubHintDomain struct {
	gotTourId uuid.UUID
	created   *domains.Hint
	err       error
}

func (s *stubHintDomain) Create(_ context.Context, tourId uuid.UUID, h *domains.Hint) (*domains.Hint, error) {
	if s.err != nil {
		return nil, s.err
	}
	s.gotTourId = tourId
	h.Id = uuid.Must(uuid.NewV7())
	h.TourId = tourId
	s.created = h
	return h, nil
}

func hintRouter(domain domains.HintDomainInterface) *gin.Engine {
	controller := controllers.NewHintController(domain)
	return newTestRouter(func(r *gin.Engine) {
		r.POST("/tours/:tourId/hints", controller.Create)
	})
}

func validHint() map[string]any {
	return map[string]any{
		"title":     "Шаг 1",
		"content":   "Нажми сюда",
		"placement": "center",
	}
}

func TestHintController_Create(t *testing.T) {
	t.Run("valid hint returns 201 with Location and no body", func(t *testing.T) {
		domain := &stubHintDomain{}
		tourId := uuid.Must(uuid.NewV7())

		rec := doJSON(t, hintRouter(domain), http.MethodPost, "/tours/"+tourId.String()+"/hints", validHint())

		require.Equal(t, http.StatusCreated, rec.Code)
		require.NotNil(t, domain.created)
		// The path parameter must reach the domain; this is the part that silently
		// broke when the handlers moved from net/http to gin.
		assert.Equal(t, tourId, domain.gotTourId)
		assert.Equal(t,
			"/v1/tours/"+tourId.String()+"/hints/"+domain.created.Id.String(),
			rec.Header().Get("Location"),
		)
		assert.Empty(t, rec.Body.String())
	})

	t.Run("non-uuid tour id returns 400", func(t *testing.T) {
		rec := doJSON(t, hintRouter(&stubHintDomain{}), http.MethodPost, "/tours/not-a-uuid/hints", validHint())

		require.Equal(t, http.StatusBadRequest, rec.Code)
		payload := decodeHTTPError(t, rec)
		assert.Contains(t, payload.Error, "URI")
	})

	t.Run("malformed JSON returns 400", func(t *testing.T) {
		tourId := uuid.Must(uuid.NewV7())
		rec := doRaw(hintRouter(&stubHintDomain{}), http.MethodPost, "/tours/"+tourId.String()+"/hints", "{oops")

		require.Equal(t, http.StatusBadRequest, rec.Code)
		decodeHTTPError(t, rec)
	})

	t.Run("missing tour returns 404", func(t *testing.T) {
		domain := &stubHintDomain{err: &domains.RepositoryError{
			Err:       domains.ErrTourNotFound,
			EntityRef: "abc",
			Operation: domains.Retrieve,
			Reason:    domains.NoRecord,
		}}
		tourId := uuid.Must(uuid.NewV7())

		rec := doJSON(t, hintRouter(domain), http.MethodPost, "/tours/"+tourId.String()+"/hints", validHint())

		require.Equal(t, http.StatusNotFound, rec.Code)
		decodeHTTPError(t, rec)
	})

	t.Run("published tour returns 409", func(t *testing.T) {
		domain := &stubHintDomain{err: &domains.LogicError{
			Err:   domains.ErrTourIsPublished,
			Stage: "tour edit",
			Code:  http.StatusConflict,
		}}
		tourId := uuid.Must(uuid.NewV7())

		rec := doJSON(t, hintRouter(domain), http.MethodPost, "/tours/"+tourId.String()+"/hints", validHint())

		require.Equal(t, http.StatusConflict, rec.Code)
		payload := decodeHTTPError(t, rec)
		assert.Contains(t, payload.Error, domains.ErrTourIsPublished.Error())
	})

	t.Run("hint validation error returns 400", func(t *testing.T) {
		domain := &stubHintDomain{err: &domains.LogicError{
			Err:   domains.ErrSelectorRequired,
			Stage: "hint validation",
			Code:  http.StatusBadRequest,
		}}
		tourId := uuid.Must(uuid.NewV7())

		rec := doJSON(t, hintRouter(domain), http.MethodPost, "/tours/"+tourId.String()+"/hints", validHint())

		require.Equal(t, http.StatusBadRequest, rec.Code)
		payload := decodeHTTPError(t, rec)
		assert.Contains(t, payload.Error, domains.ErrSelectorRequired.Error())
	})
}
