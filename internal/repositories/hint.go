package repositories

import (
	"context"
	"sort"
	"sync"

	"github.com/Vanady39/cluer/internal/domains"
	"github.com/google/uuid"
)

type (
	HintRepository struct {
		mu    sync.RWMutex
		hints map[uuid.UUID]*domains.Hint
	}
)

func NewHintRepository() *HintRepository {
	return &HintRepository{hints: make(map[uuid.UUID]*domains.Hint)}
}

// hintNotFound builds the standard not-found error for a hint, carrying
// domains.ErrHintNotFound so callers can still match it with errors.Is.
func hintNotFound(id uuid.UUID, op domains.RepoOperation) error {
	return &domains.RepositoryError{
		Err:       domains.ErrHintNotFound,
		EntityRef: id.String(),
		Operation: op,
		Reason:    domains.NoRecord,
	}
}

func (hr *HintRepository) CreateHint(_ context.Context, h *domains.Hint) error {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	cp := *h
	hr.hints[h.Id] = &cp
	return nil
}

func (hr *HintRepository) GetById(_ context.Context, id uuid.UUID) (*domains.Hint, error) {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	h, ok := hr.hints[id]
	if !ok {
		return nil, hintNotFound(id, domains.Retrieve)
	}
	cp := *h
	return &cp, nil
}

func (hr *HintRepository) ListByTour(_ context.Context, tourId uuid.UUID) ([]domains.Hint, error) {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	result := make([]domains.Hint, 0)
	for _, h := range hr.hints {
		if h.TourId == tourId {
			result = append(result, *h)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Step < result[j].Step
	})
	return result, nil
}

func (hr *HintRepository) UpdateHint(_ context.Context, h *domains.Hint) error {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	if _, ok := hr.hints[h.Id]; !ok {
		return hintNotFound(h.Id, domains.Update)
	}
	cp := *h
	hr.hints[h.Id] = &cp
	return nil
}

func (hr *HintRepository) DeleteHint(_ context.Context, tourId, id uuid.UUID) error {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	if h, ok := hr.hints[id]; ok && h.TourId == tourId {
		delete(hr.hints, id)
		return nil
	}
	return hintNotFound(id, domains.Delete)
}

func (hr *HintRepository) ReorderHint(_ context.Context, tourId uuid.UUID, ids []uuid.UUID) error {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	// Resolve every id before touching any of them, so a bad id cannot leave the
	// tour half-renumbered.
	ordered := make([]*domains.Hint, len(ids))
	for i, id := range ids {
		h, ok := hr.hints[id]
		if !ok || h.TourId != tourId {
			return &domains.RepositoryError{
				Err:       domains.ErrReorderMismatch,
				EntityRef: id.String(),
				Operation: domains.Update,
				Reason:    domains.InvalidReference,
			}
		}
		ordered[i] = h
	}

	for i, h := range ordered {
		h.Step = i + 1
	}
	return nil
}
