package onboarding

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TourService struct {
	tours TourRepository
	hints HintRepository
}

func NewTourService(t TourRepository, h HintRepository) *TourService {
	return &TourService{t, h}
}

func (s *TourService) Create(ctx context.Context, t *Tour) (*Tour, error) {
	if err := validateTour(t); err != nil {
		return nil, err
	}
	t.Id = uuid.Must(uuid.NewV7())
	t.Status = TourDraft
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()

	if err := s.tours.CreateTour(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TourService) ListPublished(ctx context.Context, path string) ([]*Tour, error) {
	tours, err := s.tours.ListPublished(ctx, path)
	if err != nil {
		return nil, err
	}

	for _, t := range tours {
		hints, err := s.hints.ListByTour(ctx, t.Id)
		if err != nil {
			return nil, err
		}
		t.Hints = hints
	}

	return tours, nil
}

func validateTour(t *Tour) error {
	if t.Title == "" {
		return ErrTitleRequired
	}
	if t.TargetPath == "" {
		return ErrTargetPathRequired
	}
	if t.Priority < 0 {
		return ErrPriorityNegative
	}
	if t.TriggerType != "" && !t.TriggerType.Valid() {
		return ErrInvalidTriggerType
	}
	return nil
}
