package domains

import (
	"context"
	"errors"
	"github.com/Vanady39/cluer/platform/errs"
	"net/http"

	"github.com/google/uuid"
)

type (
	HintDomain struct {
		tours TourRepositoryInterface
		hints HintRepositoryInterface
	}

	HintDomainInterface interface {
		Create(ctx context.Context, tourId uuid.UUID, h *Hint) (*Hint, error)
		Update(ctx context.Context, tourId, hintId uuid.UUID, h *Hint) (*Hint, error)
		Delete(ctx context.Context, tourId, hintId uuid.UUID) error
		Reorder(ctx context.Context, tourId uuid.UUID, ids []uuid.UUID) ([]Hint, error)
		ListByTour(ctx context.Context, tourId uuid.UUID) ([]Hint, error)
	}

	HintRepositoryInterface interface {
		Create(ctx context.Context, h *Hint) error
		GetById(ctx context.Context, id uuid.UUID) (*Hint, error)
		ListByVersion(ctx context.Context, versionId uuid.UUID) ([]Hint, error)
		Update(ctx context.Context, h *Hint) error
		Delete(ctx context.Context, versionId, id uuid.UUID) error
		Reorder(ctx context.Context, versionId uuid.UUID, ids []uuid.UUID) error
	}
)

func NewHintDomain(tours TourRepositoryInterface, hints HintRepositoryInterface) *HintDomain {
	return &HintDomain{tours: tours, hints: hints}
}

func (hd *HintDomain) Create(ctx context.Context, tourId uuid.UUID, h *Hint) (*Hint, error) {
	draft, err := hd.draftOf(ctx, tourId)
	if err != nil {
		return nil, err
	}

	existing, err := hd.hints.ListByVersion(ctx, draft.Id)
	if err != nil {
		return nil, err
	}

	maxStep := 0
	for _, e := range existing {
		if e.Step > maxStep {
			maxStep = e.Step
		}
	}

	h.TourVersionId = draft.Id
	h.Step = maxStep + 1

	if err := hd.hints.Create(ctx, h); err != nil {
		return nil, err
	}
	return h, nil
}

func (hd *HintDomain) Update(ctx context.Context, tourId, hintId uuid.UUID, h *Hint) (*Hint, error) {
	draft, err := hd.draftOf(ctx, tourId)
	if err != nil {
		return nil, err
	}

	current, err := hd.hints.GetById(ctx, hintId)
	if err != nil {
		return nil, err
	}
	if current.TourVersionId != draft.Id {
		return nil, logicErr(ErrVersionImmutable, "hint update", http.StatusConflict)
	}

	h.Id = hintId
	h.TourVersionId = draft.Id
	h.Step = current.Step

	if err := hd.hints.Update(ctx, h); err != nil {
		return nil, err
	}
	return h, nil
}

func (hd *HintDomain) Delete(ctx context.Context, tourId, hintId uuid.UUID) error {
	draft, err := hd.draftOf(ctx, tourId)
	if err != nil {
		return err
	}
	return hd.hints.Delete(ctx, draft.Id, hintId)
}

func (hd *HintDomain) Reorder(ctx context.Context, tourId uuid.UUID, ids []uuid.UUID) ([]Hint, error) {
	draft, err := hd.draftOf(ctx, tourId)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, logicErr(ErrReorderMismatch, "hint reorder", http.StatusBadRequest)
	}
	// Репозиторий сверяет только count(*) == len(ids), поэтому список с повтором
	// проходит его проверку, но раздаёт двум подсказкам один и тот же step.
	seen := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		if _, dup := seen[id]; dup {
			return nil, logicErr(ErrReorderMismatch, "hint reorder", http.StatusBadRequest)
		}
		seen[id] = struct{}{}
	}
	if err := hd.hints.Reorder(ctx, draft.Id, ids); err != nil {
		// Наружу должен уходить читаемый текст сентинела, а не «failed to update
		// entity <id>: invref» из RepositoryError.
		if errors.Is(err, ErrReorderMismatch) {
			return nil, logicErr(ErrReorderMismatch, "hint reorder", http.StatusBadRequest)
		}
		return nil, err
	}
	return hd.hints.ListByVersion(ctx, draft.Id)
}

func (hd *HintDomain) ListByTour(ctx context.Context, tourId uuid.UUID) ([]Hint, error) {
	draft, err := hd.draftOf(ctx, tourId)
	if err != nil {
		return nil, err
	}
	return hd.hints.ListByVersion(ctx, draft.Id)
}

func (hd *HintDomain) draftOf(ctx context.Context, tourId uuid.UUID) (*TourVersion, error) {
	draft, err := hd.tours.VersionByStatus(ctx, tourId, TourDraft)
	if err != nil {
		if errs.IsNotFound(err) {
			return nil, logicErr(ErrNoDraft, "hint edit", http.StatusConflict)
		}
		return nil, err
	}
	return draft, nil
}
