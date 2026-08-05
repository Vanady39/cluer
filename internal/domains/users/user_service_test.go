package users

import (
	"context"
	"errors"
	"testing"
)

type userRepositoryStub struct {
	user User
	err  error
}

func (s userRepositoryStub) GetCurrentUser(context.Context) (User, error) {
	return s.user, s.err
}

func TestUserServiceGetCurrentUser(t *testing.T) {
	expected := User{
		ID:        1,
		Name:      "Иванов Иван",
		AvatarURL: "https://example.com/avatar.png",
	}
	service := NewUserService(userRepositoryStub{user: expected})

	actual, err := service.GetCurrentUser(context.Background())
	if err != nil {
		t.Fatalf("GetCurrentUser() returned an unexpected error: %v", err)
	}

	if actual != expected {
		t.Fatalf("GetCurrentUser() = %#v, want %#v", actual, expected)
	}
}

func TestUserServiceGetCurrentUserReturnsRepositoryError(t *testing.T) {
	expectedErr := errors.New("repository error")
	service := NewUserService(userRepositoryStub{err: expectedErr})

	_, err := service.GetCurrentUser(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("GetCurrentUser() error = %v, want %v", err, expectedErr)
	}
}
