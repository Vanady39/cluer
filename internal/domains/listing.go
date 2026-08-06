package domains

import "context"

type (
	ListingDomain struct {
		repo ListingRepositoryInterface
	}

	ListingDomainInterface interface {
		GetListings(context.Context) ([]Listing, error)
	}

	ListingRepositoryInterface interface {
		GetListings(context.Context) ([]Listing, error)
	}
)

func NewListingDomain(repo ListingRepositoryInterface) *ListingDomain {
	return &ListingDomain{repo: repo}
}

func (ld *ListingDomain) GetListings(ctx context.Context) ([]Listing, error) {
	return ld.repo.GetListings(ctx)
}
