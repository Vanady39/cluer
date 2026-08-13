package controllers

import (
	"errors"
	"fmt"

	"github.com/Vanady39/cluer/platform/errs"
)

var (
	errMissingPath   = errors.New("path query parameter is required")
	errMissingAppId  = errors.New("appId query parameter is required")
	errBadTimeFormat = errors.New("from and to must be RFC3339 timestamps")
	errMissingAppKey = errors.New("X-App-Key header is required")
)

// -------------------------------- //
// ----- Business Logic Error ----- //
// -------------------------------- //

// LogicError is the controller-level counterpart of the domain's own: it
// reports a failure the handler itself decided on, without leaking the
// underlying error to the client.
type (
	LogicError struct {
		Err   error
		Stage string
		Code  int
	}
)

func (le *LogicError) Error() string {
	return fmt.Sprintf("failure during %s: %v", le.Stage, le.Err)
}

func (le *LogicError) Message() string {
	return fmt.Sprintf("An error occurred during %s", le.Stage)
}

func (le *LogicError) ToHTTPError() *errs.HTTPError {
	return &errs.HTTPError{Message: le.Message(), Error: fmt.Sprintf("failure during %s: %v", le.Stage, le.Err)}
}

func (le *LogicError) Unwrap() error {
	return le.Err
}

func (le *LogicError) StatusCode() int {
	return le.Code
}
