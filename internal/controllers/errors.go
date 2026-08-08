package controllers

import (
	"errors"
	"fmt"

	"github.com/Vanady39/cluer/internal/models"
)

type (
	BaseError interface {
		Error() string
		Unwrap() error
		Message() string
		ToHTTPError() *models.HTTPError
		StatusCode() int
	}
)

var (
	errMissingPath   = errors.New("path query parameter is required")
	errMissingAppId  = errors.New("appId query parameter is required")
	errBadTimeFormat = errors.New("from and to must be RFC3339 timestamps")
	errMissingAppKey = errors.New("X-App-Key header is required")
)

// ----------------------------- //
// ----- Permissions Error ----- //
// ----------------------------- //

type (
	PermissionError struct {
		Err  error
		Code int
	}
)

func (pe *PermissionError) Error() string {
	return fmt.Sprintf("access denied: %v", pe.Err)
}

func (pe *PermissionError) Message() string {
	return "Access denied"
}

func (pe *PermissionError) ToHTTPError() *models.HTTPError {
	return &models.HTTPError{Message: pe.Message(), Error: "access denied"}
}

func (pe *PermissionError) Unwrap() error {
	return pe.Err
}

func (pe *PermissionError) StatusCode() int {
	return pe.Code
}

// ------------------------- //
// ----- Binding Error ----- //
// ------------------------- //

type (
	BindingZone string

	BindingError struct {
		Err  error
		Zone BindingZone
		Code int
	}
)

const (
	URI    BindingZone = "URI"
	Body   BindingZone = "body"
	JSON   BindingZone = "JSON"
	Header BindingZone = "header"
	Query  BindingZone = "query"
)

func (be *BindingError) Error() string {
	return fmt.Sprintf("failed to bind %s: %v", be.Zone, be.Err)
}

func (be *BindingError) Message() string {
	return fmt.Sprintf("Failed to parse: invalid key/value in %s", be.Zone)
}

func (be *BindingError) ToHTTPError() *models.HTTPError {
	return &models.HTTPError{Message: be.Message(), Error: fmt.Sprintf("failed to bind %s", be.Zone)}
}

func (be *BindingError) Unwrap() error {
	return be.Err
}

func (be *BindingError) StatusCode() int {
	return be.Code
}

// -------------------------------- //
// ----- Business Logic Error ----- //
// -------------------------------- //

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

func (le *LogicError) ToHTTPError() *models.HTTPError {
	return &models.HTTPError{Message: le.Message(), Error: fmt.Sprintf("failure during %s: %v", le.Stage, le.Err)}
}

func (le *LogicError) Unwrap() error {
	return le.Err
}

func (le *LogicError) StatusCode() int {
	return le.Code
}
