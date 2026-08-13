package middlewares

import (
	"errors"

	"github.com/Vanady39/cluer/internal/models"
)

var (
	errMissingAuthHeader   = errors.New("missing Authorization header")
	errMalformedAuthHeader = errors.New("malformed Authorization header")
	errInvalidToken        = errors.New("invalid or expired id token")
	errBrokenAuthContext   = errors.New("oidc claims are unreadable")
	errInsufficientGroups  = errors.New("insufficient permissions")
)

// Authorization error
type (
	AuthHeaderError struct {
		Code int
		msg  string
		Err  error
	}
)

func (iah *AuthHeaderError) Error() string {
	return iah.Err.Error()
}

func (iah *AuthHeaderError) Message() string {
	return iah.msg
}

func (iah *AuthHeaderError) ToHTTPError() *models.HTTPError {
	return &models.HTTPError{
		Error:   iah.Err.Error(),
		Message: iah.msg,
	}
}

func (iah *AuthHeaderError) Unwrap() error {
	return iah.Err
}

func (iah *AuthHeaderError) StatusCode() int {
	return iah.Code
}
