package domains

import "context"

type (
	UserDomain struct {
		repo UserRepositoryInterface
	}

	UserDomainInterface interface {
		GetCurrentUser(context.Context) (User, error)
	}

	UserRepositoryInterface interface {
		GetCurrentUser(context.Context) (User, error)
	}
)

func NewUserDomain(repo UserRepositoryInterface) *UserDomain {
	return &UserDomain{repo: repo}
}

func (ud *UserDomain) GetCurrentUser(ctx context.Context) (User, error) {
	return ud.repo.GetCurrentUser(ctx)
}
