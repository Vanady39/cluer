package users

import (
	"context"

	domainusers "github.com/Vanady39/cluer/internal/domains/users"
)

type MockRepository struct {
	user domainusers.User
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		user: domainusers.User{
			ID:        1,
			Name:      "Иванов Иван",
			AvatarURL: "https://example.com/avatar.png",
		},
	}
}

func (r *MockRepository) GetCurrentUser(_ context.Context) (domainusers.User, error) {
	return r.user, nil
}
