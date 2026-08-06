package onboarding

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	domains "github.com/Vanady39/cluer/internal/domains/onboarding"

	"github.com/google/uuid"
)

type Handler struct {
	TourService *domains.TourService
	HintService *domains.HintService
}

func NewHandler(ts *domains.TourService, hs *domains.HintService) *Handler {
	return &Handler{
		TourService: ts,
		HintService: hs,
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// CreateTour godoc
//
//		@Summary      Create a new tour
//		@Description  Creates a new onboarding tour draft
//		@Tags         Tours
//		@Accept       json
//		@Produce      json
//		@Param        tour  body      domains.Tour  true  "Tour object"
//	 	@Success 	  201   "Created"
//		@Failure      400   {object}  ErrorResponse
//		@Failure      500   {object}  ErrorResponse
//		@Router       /tours [post]
func (h *Handler) CreateTour(w http.ResponseWriter, r *http.Request) {
	var tour domains.Tour
	if err := json.NewDecoder(r.Body).Decode(&tour); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}
	defer r.Body.Close()

	createdTour, err := h.TourService.Create(r.Context(), &tour)
	if err != nil {
		switch {
		case errors.Is(err, domains.ErrTitleRequired), errors.Is(err, domains.ErrTargetPathRequired),
			errors.Is(err, domains.ErrPriorityNegative), errors.Is(err, domains.ErrInvalidTriggerType):
			writeError(w, http.StatusBadRequest, err)
		default:
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	w.Header().Set("Location", "/api/v1/tours/"+createdTour.Id.String())
	w.WriteHeader(http.StatusCreated)
}

// CreateHint godoc
//
//		@Summary      Create a hint for a tour
//		@Description  Creates a new hint for a specific tour
//		@Tags         Hints
//		@Accept       json
//		@Produce      json
//		@Param        tourId  path    string         true  "Tour ID"  format(uuid)
//		@Param        hint    body    domains.Hint   true  "Hint object"
//	 @Success 	  201     "Created"
//		@Failure      400     {object}  ErrorResponse
//		@Failure      404     {object}  ErrorResponse
//		@Failure      409     {object}  ErrorResponse
//		@Failure      500     {object}  ErrorResponse
//		@Router       /tours/{tourId}/hints [post]
func (h *Handler) CreateHint(w http.ResponseWriter, r *http.Request) {
	tourIdStr := r.PathValue("tourId")
	tourId, err := uuid.Parse(tourIdStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid tour ID"))
		return
	}

	var hint domains.Hint
	if err := json.NewDecoder(r.Body).Decode(&hint); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}
	defer r.Body.Close()

	createdHint, err := h.HintService.Create(r.Context(), tourId, &hint)
	if err != nil {
		switch {
		case errors.Is(err, domains.ErrTitleRequired), errors.Is(err, domains.ErrContentRequired),
			errors.Is(err, domains.ErrBadPlacement), errors.Is(err, domains.ErrSelectorRequired):
			writeError(w, http.StatusBadRequest, err)
		case errors.Is(err, domains.ErrTourNotFound):
			writeError(w, http.StatusNotFound, err)
		case errors.Is(err, domains.ErrTourIsPublished), errors.Is(err, domains.ErrTourIsArchived):
			writeError(w, http.StatusConflict, err)
		default:
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/tours/%s/hints/%s", tourId, createdHint.Id.String()))
	w.WriteHeader(http.StatusCreated)
}

// GetPublishedTours godoc
//
//	@Summary      Get published tours by path
//	@Description  Returns published tours with hints for a specific target path
//	@Tags         Tours
//	@Produce      json
//	@Param        path  query     string  true  "Target path"  example(/dashboard)
//	@Success      200   {array}   domains.Tour
//	@Failure      400   {object}  ErrorResponse
//	@Failure      500   {object}  ErrorResponse
//	@Router       /tours/published [get]
func (h *Handler) GetPublishedTours(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		writeError(w, http.StatusBadRequest, errors.New("path query parameter is required"))
		return
	}

	tours, err := h.TourService.ListPublished(r.Context(), path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, tours)
}
