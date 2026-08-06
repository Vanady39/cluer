package domains

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type (
	TourDomain struct {
		tours TourRepositoryInterface
		hints HintRepositoryInterface
	}

	TourDomainInterface interface {
		Create(context.Context, *Tour) (*Tour, error)
		ListPublished(context.Context, string) ([]*Tour, error)
	}

	TourRepositoryInterface interface {
		CreateTour(ctx context.Context, t *Tour) error
		GetById(ctx context.Context, id uuid.UUID) (*Tour, error)
		ListTour(ctx context.Context) ([]*Tour, error)
		UpdateTour(ctx context.Context, tour *Tour) error
		DeleteTour(ctx context.Context, id uuid.UUID) error
		SetStatus(ctx context.Context, id uuid.UUID, s TourStatus) error
		ListPublished(ctx context.Context, path string) ([]*Tour, error)
	}
)

func NewTourDomain(tours TourRepositoryInterface, hints HintRepositoryInterface) *TourDomain {
	return &TourDomain{tours: tours, hints: hints}
}

func (td *TourDomain) Create(ctx context.Context, t *Tour) (*Tour, error) {
	if err := validateTour(t); err != nil {
		return nil, err
	}

	t.Id = uuid.Must(uuid.NewV7())
	t.Status = TourDraft
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()

	if err := td.tours.CreateTour(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (td *TourDomain) ListPublished(ctx context.Context, path string) ([]*Tour, error) {
	tours, err := td.tours.ListPublished(ctx, path)
	if err != nil {
		return nil, err
	}

	for _, t := range tours {
		hints, err := td.hints.ListByTour(ctx, t.Id)
		if err != nil {
			return nil, err
		}
		t.Hints = hints
	}

	return tours, nil
}

func validateTour(t *Tour) error {
	var err error
	switch {
	case t.Title == "":
		err = ErrTitleRequired
	case t.TargetPath == "":
		err = ErrTargetPathRequired
	case t.Priority < 0:
		err = ErrPriorityNegative
	case t.TriggerType != "" && !t.TriggerType.Valid():
		err = ErrInvalidTriggerType
	default:
		return nil
	}
	return &LogicError{Err: err, Stage: "tour validation", Code: http.StatusBadRequest}
}
