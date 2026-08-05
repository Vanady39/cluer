package users

import (
	"context"

	domainusers "github.com/Vanady39/cluer/internal/domains/users"
)

type UserRepository struct {
	user domainusers.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		user: domainusers.User{
			ID:        1,
			Name:      "Иванов Иван",
			AvatarURL: "https://example.com/avatar.png",
		},
	}
}

func (r *UserRepository) GetCurrentUser(_ context.Context) (domainusers.User, error) {
	return r.user, nil
}
