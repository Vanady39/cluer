package repositories

import (
	"context"

	"github.com/Vanady39/cluer/internal/domains"
)

type (
	ListingRepository struct {
		listings []domains.Listing
	}
)

func NewListingRepository() *ListingRepository {
	return &ListingRepository{
		listings: []domains.Listing{
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

func (lr *ListingRepository) GetListings(_ context.Context) ([]domains.Listing, error) {
	listings := make([]domains.Listing, len(lr.listings))
	copy(listings, lr.listings)

	return listings, nil
}
