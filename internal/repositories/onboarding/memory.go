package onboarding

import (
	"context"
	"sort"
	"sync"

	domain "github.com/Vanady39/cluer/internal/domains/onboarding"
	"github.com/google/uuid"
)

type MemoryHintRepository struct {
	mu    sync.RWMutex
	hints map[uuid.UUID]*domain.Hint
}

func NewMemoryHintRepository() *MemoryHintRepository {
	return &MemoryHintRepository{hints: make(map[uuid.UUID]*domain.Hint)}
}

func (r *MemoryHintRepository) CreateHint(_ context.Context, h *domain.Hint) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *h
	r.hints[h.Id] = &cp
	return nil
}

func (r *MemoryHintRepository) GetById(_ context.Context, id uuid.UUID) (*domain.Hint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if h, ok := r.hints[id]; ok {
		return h, nil
	}
	return nil, nil
}

func (r *MemoryHintRepository) ListByTour(_ context.Context, tourId uuid.UUID) ([]domain.Hint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]domain.Hint, 0)
	for _, h := range r.hints {
		if h.TourId == tourId {
			result = append(result, *h)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Step < result[j].Step
	})
	return result, nil
}

func (r *MemoryHintRepository) UpdateHint(_ context.Context, h *domain.Hint) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.hints[h.Id]; !ok {
		return domain.ErrHintNotFound
	}
	cp := *h
	r.hints[h.Id] = &cp
	return nil
}

func (r *MemoryHintRepository) DeleteHint(_ context.Context, tourId, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if h, ok := r.hints[id]; ok && h.TourId == tourId {
		delete(r.hints, id)
		return nil
	}
	return domain.ErrHintNotFound
}

func (r *MemoryHintRepository) ReorderHint(_ context.Context, tourId uuid.UUID, ids []uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, id := range ids {
		if h, ok := r.hints[id]; ok && h.TourId == tourId {
			h.Step = i + 1
		}
	}
	return nil
}

type MemoryTourRepository struct {
	mu    sync.RWMutex
	tours map[uuid.UUID]*domain.Tour
}

func NewMemoryTourRepository() *MemoryTourRepository {
	return &MemoryTourRepository{tours: make(map[uuid.UUID]*domain.Tour)}
}

func (r *MemoryTourRepository) CreateTour(_ context.Context, t *domain.Tour) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *t
	r.tours[t.Id] = &cp
	return nil
}

func (r *MemoryTourRepository) GetById(_ context.Context, id uuid.UUID) (*domain.Tour, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if t, ok := r.tours[id]; ok {
		return t, nil
	}
	return nil, nil
}

func (r *MemoryTourRepository) ListTour(_ context.Context) ([]*domain.Tour, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Tour
	for _, t := range r.tours {
		result = append(result, t)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result, nil
}

func (r *MemoryTourRepository) UpdateTour(_ context.Context, tour *domain.Tour) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tours[tour.Id]; !ok {
		return domain.ErrTourNotFound
	}
	cp := *tour
	r.tours[tour.Id] = &cp
	return nil
}

func (r *MemoryTourRepository) DeleteTour(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tours[id]; !ok {
		return domain.ErrTourNotFound
	}
	delete(r.tours, id)
	return nil
}

func (r *MemoryTourRepository) SetStatus(_ context.Context, id uuid.UUID, s domain.TourStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t, ok := r.tours[id]; ok {
		t.Status = s
		return nil
	}
	return domain.ErrTourNotFound
}

func (r *MemoryTourRepository) ListPublished(_ context.Context, path string) ([]*domain.Tour, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Tour
	for _, t := range r.tours {
		if t.Status == domain.TourPublished && t.TargetPath == path {
			cp := *t
			result = append(result, &cp)
		}
	}
	return result, nil
}
