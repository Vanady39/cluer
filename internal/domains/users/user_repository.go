package users

import "context"

type UserRepository interface {
	GetCurrentUser(ctx context.Context) (User, error)
}
