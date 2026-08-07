package repositories

import (
	"context"
	"sort"
	"sync"

	"github.com/Vanady39/cluer/internal/domains"
	"github.com/google/uuid"
)

type (
	TourRepository struct {
		mu    sync.RWMutex
		tours map[uuid.UUID]*domains.Tour
	}
)

func NewTourRepository() *TourRepository {
	return &TourRepository{tours: make(map[uuid.UUID]*domains.Tour)}
}

// tourNotFound builds the standard not-found error for a tour, carrying
// domains.ErrTourNotFound so callers can still match it with errors.Is.
func tourNotFound(id uuid.UUID) error {
	return &domains.RepositoryError{
		Err:       domains.ErrTourNotFound,
		EntityRef: id.String(),
		Operation: domains.Retrieve,
		Reason:    domains.NoRecord,
	}
}

func (tr *TourRepository) CreateTour(_ context.Context, t *domains.Tour) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	cp := *t
	tr.tours[t.Id] = &cp
	return nil
}

func (tr *TourRepository) GetById(_ context.Context, id uuid.UUID) (*domains.Tour, error) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	t, ok := tr.tours[id]
	if !ok {
		return nil, tourNotFound(id)
	}
	cp := *t
	return &cp, nil
}

func (tr *TourRepository) ListTour(_ context.Context) ([]*domains.Tour, error) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	result := make([]*domains.Tour, 0, len(tr.tours))
	for _, t := range tr.tours {
		cp := *t
		result = append(result, &cp)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result, nil
}

func (tr *TourRepository) UpdateTour(_ context.Context, tour *domains.Tour) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	if _, ok := tr.tours[tour.Id]; !ok {
		return tourNotFound(tour.Id)
	}
	cp := *tour
	tr.tours[tour.Id] = &cp
	return nil
}

func (tr *TourRepository) DeleteTour(_ context.Context, id uuid.UUID) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	if _, ok := tr.tours[id]; !ok {
		return tourNotFound(id)
	}
	delete(tr.tours, id)
	return nil
}

func (tr *TourRepository) SetStatus(_ context.Context, id uuid.UUID, s domains.TourStatus) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	t, ok := tr.tours[id]
	if !ok {
		return tourNotFound(id)
	}
	t.Status = s
	return nil
}

func (tr *TourRepository) ListPublished(_ context.Context, path string) ([]*domains.Tour, error) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	result := make([]*domains.Tour, 0)
	for _, t := range tr.tours {
		if t.Status == domains.TourPublished && t.TargetPath == path {
			cp := *t
			result = append(result, &cp)
		}
	}
	return result, nil
}
