package repositories

import (
	"context"

	"github.com/Vanady39/cluer/internal/domains"
)

type (
	UserRepository struct {
		user domains.User
	}
)

func NewUserRepository() *UserRepository {
	return &UserRepository{
		user: domains.User{
			ID:        1,
			Name:      "Иванов Иван",
			AvatarURL: "https://example.com/avatar.png",
		},
	}
}

func (ur *UserRepository) GetCurrentUser(_ context.Context) (domains.User, error) {
	return ur.user, nil
}
