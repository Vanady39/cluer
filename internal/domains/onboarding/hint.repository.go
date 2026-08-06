package onboarding

import (
	"context"

	"github.com/google/uuid"
)

type HintRepository interface {
	CreateHint(ctx context.Context, h *Hint) error
	GetById(ctx context.Context, id uuid.UUID) (*Hint, error)
	ListByTour(ctx context.Context, tourId uuid.UUID) ([]Hint, error)
	UpdateHint(ctx context.Context, h *Hint) error
	DeleteHint(ctx context.Context, tourId, id uuid.UUID) error
	ReorderHint(ctx context.Context, tourId uuid.UUID, ids []uuid.UUID) error
}
