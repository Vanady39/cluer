package domains

type Listing struct {
	ID          int64
	Title       string
	Description string
	Price       int64
	ImageURL    string
}

const (
	DefaultListingsLimit  = 20
	MaxListingsLimit      = 50
	MaxListingsOffset     = 100000
	MaxListingsQueryRunes = 200
)

type ListingFilter struct {
	Query  string
	Limit  int
	Offset int
}

type ListingPage struct {
	Listings []Listing
	Total    int64
}

// NormalizeListingFilter keeps repository callers safe when they do not come
// through the HTTP controller, such as background jobs and tests.
func NormalizeListingFilter(filter ListingFilter) ListingFilter {
	if filter.Limit <= 0 {
		filter.Limit = DefaultListingsLimit
	} else if filter.Limit > MaxListingsLimit {
		filter.Limit = MaxListingsLimit
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	} else if filter.Offset > MaxListingsOffset {
		filter.Offset = MaxListingsOffset
	}
	return filter
}
