package users

import (
	"context"

	domainusers "github.com/Vanady39/cluer/internal/domains/users"
)

type Repository interface {
	GetCurrentUser(ctx context.Context) (domainusers.User, error)
}

type Service struct {
	repository Repository
}

func New(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) GetCurrentUser(ctx context.Context) (domainusers.User, error) {
	return s.repository.GetCurrentUser(ctx)
}
