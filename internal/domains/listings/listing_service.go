package listings

import "context"

type ListingService interface {
	GetListings(ctx context.Context) ([]Listing, error)
}

type listingService struct {
	repository ListingRepository
}

func NewListingService(repository ListingRepository) ListingService {
	return &listingService{repository: repository}
}

func (s *listingService) GetListings(ctx context.Context) ([]Listing, error) {
	return s.repository.GetListings(ctx)
}