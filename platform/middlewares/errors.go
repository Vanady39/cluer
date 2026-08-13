package middlewares

import (
	"errors"

	"github.com/Vanady39/cluer/platform/errs"
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

// NewAuthHeaderError builds the error from outside this package. The client
// message is deliberately a separate argument from Err: the wrapped error is
// for logs, the message is what the caller is allowed to read, and keeping the
// field unexported is what stops the two from being confused.
func NewAuthHeaderError(code int, msg string, err error) *AuthHeaderError {
	return &AuthHeaderError{Code: code, msg: msg, Err: err}
}

func (iah *AuthHeaderError) Error() string {
	return iah.Err.Error()
}

func (iah *AuthHeaderError) Message() string {
	return iah.msg
}

func (iah *AuthHeaderError) ToHTTPError() *errs.HTTPError {
	return &errs.HTTPError{
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
