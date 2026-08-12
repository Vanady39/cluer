package repositories

import (
	"context"
	"strings"

	"github.com/Vanady39/cluer/internal/domains"
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
		WITH filtered AS (
			SELECT id, title, description, price, image_url
			FROM listings
			WHERE $1 = ''
			   OR title ILIKE '%' || $1 || '%' ESCAPE '\'
			   OR description ILIKE '%' || $1 || '%' ESCAPE '\'
		), paged AS (
			SELECT id, title, description, price, image_url
			FROM filtered
			ORDER BY id DESC
			LIMIT $2 OFFSET $3
		)
		SELECT (SELECT COUNT(*) FROM filtered), p.id, p.title, p.description, p.price, p.image_url
		FROM (SELECT 1) AS anchor
		LEFT JOIN paged p ON TRUE
		ORDER BY p.id DESC NULLS LAST`, query, filter.Limit, filter.Offset)
	if err != nil {
		return domains.ListingPage{}, &domains.RepositoryError{
			Err:       err,
			EntityRef: "listings",
			Operation: domains.Retrieve,
			Reason:    domains.QueryError,
		}
	}
	defer rows.Close()

	page := domains.ListingPage{Listings: make([]domains.Listing, 0)}
	for rows.Next() {
		var (
			listing     domains.Listing
			id          *int64
			title       *string
			description *string
			price       *int64
			imageURL    *string
		)
		if err := rows.Scan(
			&page.Total,
			&id,
			&title,
			&description,
			&price,
			&imageURL,
		); err != nil {
			return domains.ListingPage{}, &domains.RepositoryError{
				Err:       err,
				EntityRef: "listings",
				Operation: domains.Retrieve,
				Reason:    domains.QueryError,
			}
		}
		if id != nil {
			listing.ID = *id
			listing.Title = *title
			listing.Description = *description
			listing.Price = *price
			listing.ImageURL = *imageURL
			page.Listings = append(page.Listings, listing)
		}
	}

	if err := rows.Err(); err != nil {
		return domains.ListingPage{}, &domains.RepositoryError{
			Err:       err,
			EntityRef: "listings",
			Operation: domains.Retrieve,
			Reason:    domains.QueryError,
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
