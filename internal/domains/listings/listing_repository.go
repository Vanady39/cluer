package listings

import "context"

type ListingRepository interface {
	GetListings(ctx context.Context) ([]Listing, error)
}