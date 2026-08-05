package users

import "context"

type UserService interface {
	GetCurrentUser(ctx context.Context) (User, error)
}

type userService struct {
	repository UserRepository
}

func NewUserService(repository UserRepository) UserService {
	return &userService{repository: repository}
}

func (s *userService) GetCurrentUser(ctx context.Context) (User, error) {
	return s.repository.GetCurrentUser(ctx)
}
