package domains

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type (
	HintDomain struct {
		tours TourRepositoryInterface
		hints HintRepositoryInterface
	}

	HintDomainInterface interface {
		Create(context.Context, uuid.UUID, *Hint) (*Hint, error)
	}

	HintRepositoryInterface interface {
		CreateHint(ctx context.Context, h *Hint) error
		GetById(ctx context.Context, id uuid.UUID) (*Hint, error)
		ListByTour(ctx context.Context, tourId uuid.UUID) ([]Hint, error)
		UpdateHint(ctx context.Context, h *Hint) error
		DeleteHint(ctx context.Context, tourId, id uuid.UUID) error
		ReorderHint(ctx context.Context, tourId uuid.UUID, ids []uuid.UUID) error
	}
)

func NewHintDomain(tours TourRepositoryInterface, hints HintRepositoryInterface) *HintDomain {
	return &HintDomain{tours: tours, hints: hints}
}

func (hd *HintDomain) Create(ctx context.Context, tourId uuid.UUID, h *Hint) (*Hint, error) {
	if err := hd.ensureTourEditable(ctx, tourId); err != nil {
		return nil, err
	}
	if err := validateHint(h); err != nil {
		return nil, err
	}

	existing, err := hd.hints.ListByTour(ctx, tourId)
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

	if err := hd.hints.CreateHint(ctx, h); err != nil {
		return nil, err
	}
	return h, nil
}

func (hd *HintDomain) ensureTourEditable(ctx context.Context, tourId uuid.UUID) error {
	tour, err := hd.tours.GetById(ctx, tourId)
	if err != nil {
		return err
	}
	if tour == nil {
		return &RepositoryError{
			Err:       ErrTourNotFound,
			EntityRef: tourId.String(),
			Operation: Retrieve,
			Reason:    NoRecord,
		}
	}

	var reason error
	switch tour.Status {
	case TourPublished:
		reason = ErrTourIsPublished
	case TourArchived:
		reason = ErrTourIsArchived
	default:
		return nil
	}
	return &LogicError{Err: reason, Stage: "tour edit", Code: http.StatusConflict}
}

func validateHint(h *Hint) error {
	var err error
	switch {
	case h.Title == "":
		err = ErrTitleRequired
	case h.Content == "":
		err = ErrContentRequired
	case !h.Placement.Valid():
		err = ErrBadPlacement
	case h.Placement != PlacementCenter && h.Selector == "":
		err = ErrSelectorRequired
	default:
		return nil
	}
	return &LogicError{Err: err, Stage: "hint validation", Code: http.StatusBadRequest}
}
