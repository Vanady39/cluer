package listings

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"

	domainlistings "github.com/Vanady39/cluer/internal/domains/listings"
	apiresponse "github.com/Vanady39/cluer/internal/models/response"
	listingresponse "github.com/Vanady39/cluer/internal/models/response/listings"
)

type listingServiceStub struct {
	listings []domainlistings.Listing
	err      error
}

func (s listingServiceStub) GetListings(context.Context) ([]domainlistings.Listing, error) {
	return s.listings, s.err
}

func TestListingHandlerGetListings(t *testing.T) {
	expected := []domainlistings.Listing{
		{
			ID:          1,
			Title:       "iPhone 15 Pro",
			Description: "Телефон в отличном состоянии",
			Price:       95000,
			ImageURL:    "https://example.com/iphone.png",
		},
	}
	handler := NewListingHandler(listingServiceStub{listings: expected})
	context, recorder := newGinTestContext()

	handler.GetListings(context)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var actual listingresponse.GetListingsResponse
	if err := json.NewDecoder(recorder.Body).Decode(&actual); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	expectedResponse := listingresponse.NewGetListingsResponse(expected)
	if !reflect.DeepEqual(actual, expectedResponse) {
		t.Fatalf("response = %#v, want %#v", actual, expectedResponse)
	}
}

func TestListingHandlerGetListingsReturnsInternalError(t *testing.T) {
	handler := NewListingHandler(listingServiceStub{err: errors.New("service error")})
	context, recorder := newGinTestContext()

	handler.GetListings(context)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}

	var actual apiresponse.ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&actual); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if actual.Error.Code != "INTERNAL_ERROR" {
		t.Fatalf("error code = %q, want %q", actual.Error.Code, "INTERNAL_ERROR")
	}
}

func newGinTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/v1/listings", nil)

	return context, recorder
}
