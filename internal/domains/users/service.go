package users

import "context"

type Service interface {
	GetCurrentUser(ctx context.Context) (User, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) GetCurrentUser(ctx context.Context) (User, error) {
	return s.repository.GetCurrentUser(ctx)
}
