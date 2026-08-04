package users

import "context"

type Repository interface {
	GetCurrentUser(ctx context.Context) (User, error)
}
