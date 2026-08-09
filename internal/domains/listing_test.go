package domains

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type listingRepositoryStub struct {
	listings []Listing
	err      error
	filter   ListingFilter
}

func (s *listingRepositoryStub) GetListings(_ context.Context, filter ListingFilter) ([]Listing, error) {
	s.filter = filter
	return s.listings, s.err
}

func TestListingServiceGetListings(t *testing.T) {
	expected := []Listing{
		{
			ID:          1,
			Title:       "iPhone 15 Pro",
			Description: "Телефон в отличном состоянии",
			Price:       95000,
			ImageURL:    "https://example.com/iphone.png",
		},
	}
	repo := &listingRepositoryStub{listings: expected}
	service := NewListingDomain(repo)
	filter := ListingFilter{Query: "iphone", Limit: 5, Offset: 2}

	actual, err := service.GetListings(context.Background(), filter)
	if err != nil {
		t.Fatalf("GetListings() returned an unexpected error: %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("GetListings() = %#v, want %#v", actual, expected)
	}
	if !reflect.DeepEqual(repo.filter, filter) {
		t.Fatalf("repository filter = %#v, want %#v", repo.filter, filter)
	}
}

func TestListingServiceGetListingsReturnsRepositoryError(t *testing.T) {
	expectedErr := errors.New("repository error")
	service := NewListingDomain(&listingRepositoryStub{err: expectedErr})

	_, err := service.GetListings(context.Background(), ListingFilter{Limit: 20})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("GetListings() error = %v, want %v", err, expectedErr)
	}
}
