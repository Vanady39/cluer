package repositories

import (
	"context"

	"github.com/Vanady39/cluer/internal/domains"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ListingRepository struct {
	pool *pgxpool.Pool
}

func NewListingRepository(pool *pgxpool.Pool) *ListingRepository {
	return &ListingRepository{pool: pool}
}

func (lr *ListingRepository) GetListings(ctx context.Context, filter domains.ListingFilter) ([]domains.Listing, error) {
	rows, err := lr.pool.Query(ctx, `
		SELECT id, title, description, price, image_url
		FROM listings
		WHERE $1 = ''
		   OR title ILIKE '%' || $1 || '%'
		   OR description ILIKE '%' || $1 || '%'
		ORDER BY id DESC
		LIMIT $2 OFFSET $3`, filter.Query, filter.Limit, filter.Offset)
	if err != nil {
		return nil, &domains.RepositoryError{
			Err:       err,
			EntityRef: "listings",
			Operation: domains.Retrieve,
			Reason:    domains.QueryError,
		}
	}
	defer rows.Close()

	listings := make([]domains.Listing, 0)
	for rows.Next() {
		listing := domains.Listing{}
		if err := rows.Scan(
			&listing.ID,
			&listing.Title,
			&listing.Description,
			&listing.Price,
			&listing.ImageURL,
		); err != nil {
			return nil, &domains.RepositoryError{
				Err:       err,
				EntityRef: "listings",
				Operation: domains.Retrieve,
				Reason:    domains.QueryError,
			}
		}
		listings = append(listings, listing)
	}

	if err := rows.Err(); err != nil {
		return nil, &domains.RepositoryError{
			Err:       err,
			EntityRef: "listings",
			Operation: domains.Retrieve,
			Reason:    domains.QueryError,
		}
	}

	return listings, nil
}
