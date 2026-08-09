package domains

import "context"

type (
	ListingDomain struct {
		repo ListingRepositoryInterface
	}

	ListingDomainInterface interface {
		GetListings(context.Context, ListingFilter) ([]Listing, error)
	}

	ListingRepositoryInterface interface {
		GetListings(context.Context, ListingFilter) ([]Listing, error)
	}
)

func NewListingDomain(repo ListingRepositoryInterface) *ListingDomain {
	return &ListingDomain{repo: repo}
}

func (ld *ListingDomain) GetListings(ctx context.Context, filter ListingFilter) ([]Listing, error) {
	return ld.repo.GetListings(ctx, filter)
}
