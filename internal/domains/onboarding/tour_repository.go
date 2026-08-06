package onboarding

import (
	"context"

	"github.com/google/uuid"
)

type TourRepository interface {
	CreateTour(ctx context.Context, t *Tour) error
	GetById(ctx context.Context, id uuid.UUID) (*Tour, error)
	ListTour(ctx context.Context) ([]*Tour, error)
	UpdateTour(ctx context.Context, tour *Tour) error
	DeleteTour(ctx context.Context, id uuid.UUID) error
	SetStatus(ctx context.Context, id uuid.UUID, s TourStatus) error
	ListPublished(ctx context.Context, path string) ([]*Tour, error)
}
