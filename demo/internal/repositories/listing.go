package repositories

import (
	"context"
	"github.com/Vanady39/cluer/platform/errs"
	"strings"

	"github.com/Vanady39/cluer/demo/internal/domains"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type listingQueryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

type ListingRepository struct {
	queryer listingQueryer
}

func NewListingRepository(pool *pgxpool.Pool) *ListingRepository {
	return newListingRepository(pool)
}

func newListingRepository(queryer listingQueryer) *ListingRepository {
	return &ListingRepository{queryer: queryer}
}

func (lr *ListingRepository) GetListings(ctx context.Context, filter domains.ListingFilter) (domains.ListingPage, error) {
	filter = domains.NormalizeListingFilter(filter)
	query := escapeLike(filter.Query)

	rows, err := lr.queryer.Query(ctx, `
		SELECT id, title, description, price, image_url
		FROM listings
		WHERE $1 = ''
		   OR title ILIKE '%' || $1 || '%' ESCAPE '\'
		   OR description ILIKE '%' || $1 || '%' ESCAPE '\'
		ORDER BY id DESC
		LIMIT $2 OFFSET $3`, query, filter.Limit, filter.Offset)
	if err != nil {
		return domains.ListingPage{}, &errs.RepositoryError{
			Err:       err,
			EntityRef: "listings",
			Operation: errs.Retrieve,
			Reason:    errs.QueryError,
		}
	}
	defer rows.Close()

	page := domains.ListingPage{Listings: make([]domains.Listing, 0)}
	for rows.Next() {
		var listing domains.Listing
		if err := rows.Scan(
			&listing.ID,
			&listing.Title,
			&listing.Description,
			&listing.Price,
			&listing.ImageURL,
		); err != nil {
			return domains.ListingPage{}, &errs.RepositoryError{
				Err:       err,
				EntityRef: "listings",
				Operation: errs.Retrieve,
				Reason:    errs.QueryError,
			}
		}
		page.Listings = append(page.Listings, listing)
	}

	if err := rows.Err(); err != nil {
		return domains.ListingPage{}, &errs.RepositoryError{
			Err:       err,
			EntityRef: "listings",
			Operation: errs.Retrieve,
			Reason:    errs.QueryError,
		}
	}

	return page, nil
}

func escapeLike(query string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"%", "\\%",
		"_", "\\_",
	)
	return replacer.Replace(query)
}
