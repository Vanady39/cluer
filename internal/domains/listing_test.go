package domains

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type listingRepositoryStub struct {
	page   ListingPage
	err    error
	filter ListingFilter
}

func (s *listingRepositoryStub) GetListings(_ context.Context, filter ListingFilter) (ListingPage, error) {
	s.filter = filter
	return s.page, s.err
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
	repo := &listingRepositoryStub{page: ListingPage{Listings: expected, Total: 1}}
	service := NewListingDomain(repo)
	filter := ListingFilter{Query: "iphone", Limit: 5, Offset: 2}

	actual, err := service.GetListings(context.Background(), filter)
	if err != nil {
		t.Fatalf("GetListings() returned an unexpected error: %v", err)
	}

	if !reflect.DeepEqual(actual.Listings, expected) || actual.Total != 1 {
		t.Fatalf("GetListings() = %#v, want listings %#v with total 1", actual, expected)
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

func TestListingDomainNormalizesRepositoryFilter(t *testing.T) {
	repo := &listingRepositoryStub{}
	service := NewListingDomain(repo)

	_, err := service.GetListings(context.Background(), ListingFilter{Query: "iphone", Offset: -1})
	if err != nil {
		t.Fatalf("GetListings() returned an unexpected error: %v", err)
	}

	want := ListingFilter{Query: "iphone", Limit: DefaultListingsLimit, Offset: 0}
	if !reflect.DeepEqual(repo.filter, want) {
		t.Fatalf("repository filter = %#v, want %#v", repo.filter, want)
	}

	_, err = service.GetListings(context.Background(), ListingFilter{Limit: MaxListingsLimit + 1, Offset: MaxListingsOffset + 1})
	if err != nil {
		t.Fatalf("GetListings() returned an unexpected error: %v", err)
	}
	want = ListingFilter{Limit: MaxListingsLimit, Offset: MaxListingsOffset}
	if !reflect.DeepEqual(repo.filter, want) {
		t.Fatalf("repository filter = %#v, want %#v", repo.filter, want)
	}
}
