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
}

func (s listingRepositoryStub) GetListings(context.Context) ([]Listing, error) {
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
	service := NewListingDomain(listingRepositoryStub{listings: expected})

	actual, err := service.GetListings(context.Background())
	if err != nil {
		t.Fatalf("GetListings() returned an unexpected error: %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("GetListings() = %#v, want %#v", actual, expected)
	}
}

func TestListingServiceGetListingsReturnsRepositoryError(t *testing.T) {
	expectedErr := errors.New("repository error")
	service := NewListingDomain(listingRepositoryStub{err: expectedErr})

	_, err := service.GetListings(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("GetListings() error = %v, want %v", err, expectedErr)
	}
}
