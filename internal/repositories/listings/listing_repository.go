package listings

import (
	"context"

	domainlistings "github.com/Vanady39/cluer/internal/domains/listings"
)

type ListingRepository struct {
	listings []domainlistings.Listing
}

func NewListingRepository() *ListingRepository {
	return &ListingRepository{
		listings: []domainlistings.Listing{
			{
				ID:          1,
				Title:       "iPhone 15 Pro",
				Description: "Телефон в отличном состоянии",
				Price:       95000,
				ImageURL:    "https://example.com/iphone.png",
			},
			{
				ID:          2,
				Title:       "Игровое кресло",
				Description: "Использовалось полгода",
				Price:       15000,
				ImageURL:    "https://example.com/chair.png",
			},
		},
	}
}

func (r *ListingRepository) GetListings(_ context.Context) ([]domainlistings.Listing, error) {
	listings := make([]domainlistings.Listing, len(r.listings))
	copy(listings, r.listings)

	return listings, nil
}