package response

import "github.com/Vanady39/cluer/internal/domains"

type Listing struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	ImageURL    string `json:"imageUrl"`
}

type GetListingsResponse struct {
	Data []Listing `json:"data"`
}

func NewGetListingsResponse(
	domainListings []domains.Listing,
) GetListingsResponse {
	listings := make([]Listing, 0, len(domainListings))

	for _, listing := range domainListings {
		listings = append(listings, Listing{
			ID:          listing.ID,
			Title:       listing.Title,
			Description: listing.Description,
			Price:       listing.Price,
			ImageURL:    listing.ImageURL,
		})
	}

	return GetListingsResponse{Data: listings}
}
