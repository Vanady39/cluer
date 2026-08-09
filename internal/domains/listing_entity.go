package domains

type Listing struct {
	ID          int64
	Title       string
	Description string
	Price       int64
	ImageURL    string
}

type ListingFilter struct {
	Query  string
	Limit  int
	Offset int
}
