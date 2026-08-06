package domains

import (
	"errors"
	"fmt"

	"github.com/Vanady39/cluer/internal/models"
)

// ---------------------------- //
// ----- Repository Error ----- //
// ---------------------------- //

type (
	RepoOperation string

	RepoErrorReason string

	RepositoryErrorInterface interface {
		Error() string
		Unwrap() error
		Message() string
		// GetReason() RepoErrorReason
		ToHTTPError() *models.HTTPError
		StatusCode() int
	}

	RepositoryError struct {
		Err       error
		EntityRef string
		Operation RepoOperation
		Reason    RepoErrorReason
	}
)

// Repository operations
const (
	Create   RepoOperation = "create"
	Retrieve RepoOperation = "retrieve"
	Update   RepoOperation = "update"
	Delete   RepoOperation = "delete"
)

// Repository error reasons
const (
	Unknown          RepoErrorReason = "unknown"
	NoRecord         RepoErrorReason = "norecord"
	AlreadyExists    RepoErrorReason = "exists"
	CacheMiss        RepoErrorReason = "cachemmiss"
	QueryError       RepoErrorReason = "query"
	ExecError        RepoErrorReason = "exec"
	InvalidReference RepoErrorReason = "invref"
)

func (re *RepositoryError) Error() string {
	return fmt.Sprintf("failed to %s entity %s: %v", re.Operation, re.EntityRef, re.Err)
}

func (re *RepositoryError) Message() string {
	return fmt.Sprintf("Unable to %s entity %s", re.Operation, re.EntityRef)
}

// func (re *RepositoryError) SafeError() string {
// 	return fmt.Sprintf("failed to %s entity %s: %s", re.Operation, re.EntityRef, re.Reason)
// }

// func (re *RepositoryError) GetReason() RepoErrorReason {
// 	return re.Reason
// }

func (re *RepositoryError) Unwrap() error {
	return re.Err
}

func (re *RepositoryError) StatusCode() int {
	switch re.Reason {
	case Unknown:
		return 500
	case NoRecord:
		return 404
	default:
		return 500
	}
}

func (re *RepositoryError) ToHTTPError() *models.HTTPError {
	return &models.HTTPError{
		Error:   fmt.Sprintf("failed to %s entity %s: %s", re.Operation, re.EntityRef, re.Reason),
		Message: re.Message(),
	}
}

// ----------------------------- //
// ----- Permissions Error ----- //
// ----------------------------- //

type (
	PermissionErrorReason string

	PermissionError struct {
		Err       error
		EntityRef string
		Reason    PermissionErrorReason
	}
)

const (
	Unauthorized      PermissionErrorReason = "unauthorized"
	InsufficientPerms PermissionErrorReason = "insufficient"
	Disabled          PermissionErrorReason = "disabled"
)

func (pe *PermissionError) Error() string {
	return fmt.Sprintf("access denied: %s is %s for this operation: %s", pe.EntityRef, pe.Reason, pe.Err)
}

func (pe *PermissionError) Message() string {
	return fmt.Sprintf("Access denied: %s is %s for this operation", pe.EntityRef, pe.Reason)
}

func (pe *PermissionError) Unwrap() error {
	return pe.Err
}

func (pe *PermissionError) StatusCode() int {
	switch pe.Reason {
	case Unauthorized:
		return 401
	case InsufficientPerms:
		return 403
	case Disabled:
		return 423
	default:
		return 500
	}
}

func (pe *PermissionError) ToHTTPError() *models.HTTPError {
	return &models.HTTPError{
		Error:   "access denied",
		Message: pe.Message(),
	}
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
	return &models.HTTPError{Message: le.Message(), Error: fmt.Sprintf("failure during %s: %s", le.Stage, le.Err)}
}

func (le *LogicError) Unwrap() error {
	return le.Err
}

func (le *LogicError) StatusCode() int {
	return le.Code
}

// ------------------------------ //
// ----- Onboarding Reasons ----- //
// ------------------------------ //

// Sentinels carried as the Err of a LogicError or RepositoryError, so callers can
// still match them with errors.Is through the wrapper's Unwrap.

var (
	ErrTitleRequired      = errors.New("title is required")
	ErrContentRequired    = errors.New("content is required")
	ErrTargetPathRequired = errors.New("target_path is required")
	ErrPriorityNegative   = errors.New("priority must be >= 0")
)

var (
	ErrTourNotFound         = errors.New("tour not found")
	ErrTourAlreadyPublished = errors.New("tour is already published")
	ErrTourAlreadyArchived  = errors.New("tour is already archived")
	ErrTourIsDraft          = errors.New("tour is draft")
	ErrTourIsPublished      = errors.New("tour is published: editing is forbidden")
	ErrTourIsArchived       = errors.New("tour is archived: editing is forbidden")
	ErrTourHasNoHints       = errors.New("tour has no hints: cannot publish")
	ErrInvalidStatusChange  = errors.New("invalid status transition")
	ErrInvalidTriggerType   = errors.New("invalid trigger type")
)

var (
	ErrHintNotFound     = errors.New("hint not found")
	ErrBadPlacement     = errors.New("invalid placement value")
	ErrBadActionType    = errors.New("invalid action type")
	ErrSelectorRequired = errors.New("selector is required when placement is not center")
	ErrHintNotInTour    = errors.New("hint does not belong to the specified tour")
	ErrReorderMismatch  = errors.New("reorder ids do not match existing tour hints")
)
