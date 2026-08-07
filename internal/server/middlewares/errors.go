package middlewares

import (
	"errors"

	"github.com/Vanady39/cluer/internal/models"
)

var errUnauthorized = errors.New("missing or invalid admin credentials")

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
