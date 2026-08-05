package users

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	domainusers "github.com/Vanady39/cluer/internal/domains/users"
	apiresponse "github.com/Vanady39/cluer/internal/models/response"
	userresponse "github.com/Vanady39/cluer/internal/models/response/users"
)

type userServiceStub struct {
	user domainusers.User
	err  error
}

func (s userServiceStub) GetCurrentUser(context.Context) (domainusers.User, error) {
	return s.user, s.err
}

func TestUserHandlerGetCurrentUser(t *testing.T) {
	expected := domainusers.User{
		ID:        1,
		Name:      "Иванов Иван",
		AvatarURL: "https://example.com/avatar.png",
	}
	handler := NewUserHandler(userServiceStub{user: expected})
	ginContext, recorder := newUserGinTestContext()

	handler.GetCurrentUser(ginContext)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var actual userresponse.GetCurrentUserResponse
	if err := json.NewDecoder(recorder.Body).Decode(&actual); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	expectedResponse := userresponse.NewGetCurrentUserResponse(expected)
	if actual != expectedResponse {
		t.Fatalf("response = %#v, want %#v", actual, expectedResponse)
	}
}

func TestUserHandlerGetCurrentUserReturnsInternalError(t *testing.T) {
	handler := NewUserHandler(userServiceStub{err: errors.New("service error")})
	ginContext, recorder := newUserGinTestContext()

	handler.GetCurrentUser(ginContext)

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

func newUserGinTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)

	return ginContext, recorder
}
