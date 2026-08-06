package onboarding

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type HintService struct {
	tours TourRepository
	hints HintRepository
}

func NewHintService(t TourRepository, h HintRepository) *HintService {
	return &HintService{tours: t, hints: h}
}

func (s *HintService) Create(ctx context.Context, tourId uuid.UUID, h *Hint) (*Hint, error) {
	if _, err := s.ensureTourEditable(ctx, tourId); err != nil {
		return nil, err
	}
	if err := validateHint(h); err != nil {
		return nil, err
	}

	existing, err := s.hints.ListByTour(ctx, tourId)
	if err != nil {
		return nil, err
	}

	maxStep := 0
	for _, e := range existing {
		if e.Step > maxStep {
			maxStep = e.Step
		}
	}

	h.Id = uuid.Must(uuid.NewV7())
	h.TourId = tourId
	h.Step = maxStep + 1
	h.CreatedAt = time.Now()
	h.UpdatedAt = time.Now()

	if err := s.hints.CreateHint(ctx, h); err != nil {
		return nil, err
	}
	return h, nil
}

func validateHint(h *Hint) error {
	if h.Title == "" {
		return ErrTitleRequired
	}
	if h.Content == "" {
		return ErrContentRequired
	}
	if !h.Placement.Valid() {
		return ErrBadPlacement
	}
	if h.Placement != PlacementCenter && h.Selector == "" {
		return ErrSelectorRequired
	}
	return nil
}

func (s *HintService) ensureTourEditable(ctx context.Context, tourId uuid.UUID) (*Tour, error) {
	tour, err := s.tours.GetById(ctx, tourId)
	if err != nil {
		return nil, err
	}
	if tour == nil {
		return nil, ErrTourNotFound
	}
	switch tour.Status {
	case TourPublished:
		return nil, ErrTourIsPublished
	case TourArchived:
		return nil, ErrTourIsArchived
	}
	return tour, nil
}
