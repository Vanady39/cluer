package domains

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

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
	if re.Reason == NoRecord && re.Err != nil {
		return capitalise(re.Err.Error())
	}
	return fmt.Sprintf("Unable to %s the requested entity", re.Operation)
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
	case NoRecord:
		return http.StatusNotFound
	case AlreadyExists:
		return http.StatusConflict
	case InvalidReference:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
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
	if le.Err == nil {
		return fmt.Sprintf("An error occurred during %s", le.Stage)
	}
	return capitalise(le.Err.Error())
}

func (le *LogicError) ToHTTPError() *models.HTTPError {
	return &models.HTTPError{Message: le.Message(), Error: fmt.Sprintf("failure during %s: %s", le.Stage, le.Err)}
}

func capitalise(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	if r[0] >= 'a' && r[0] <= 'z' {
		r[0] -= 'a' - 'A'
	}
	return string(r)
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

var (
	ErrVersionNotFound     = errors.New("version not found")
	ErrVersionNotInTour    = errors.New("version does not belong to the specified tour")
	ErrVersionImmutable    = errors.New("published and archived versions cannot be edited")
	ErrNoDraft             = errors.New("tour has no draft to work with")
	ErrDraftAlreadyExists  = errors.New("tour already has a draft")
	ErrDraftBlocksRollback = errors.New("publish or discard the existing draft before rolling back")
	ErrNoVersionToCopy     = errors.New("tour has no published version to copy")
	ErrTourNotPublishable  = errors.New("tour cannot be published")
	ErrMaxShowsNegative    = errors.New("audience.max_shows must be >= 0")
	ErrDraftInvalid        = errors.New("draft is invalid")
)

var (
	ErrAppNotFound      = errors.New("app not found")
	ErrAppKeyRequired   = errors.New("X-App-Key header is required")
	ErrOriginNotAllowed = errors.New("origin is not in the allowed list")
	ErrSubjectRequired  = errors.New("subjectId is required")
	ErrSessionRequired  = errors.New("sessionId is required")
	ErrBatchTooLarge    = errors.New("too many events in one batch")
	ErrEmptyBatch       = errors.New("events must not be empty")
	ErrUnknownEventType = errors.New("unknown event type")
	ErrHintIdMismatch   = errors.New("hint id must be set for hint-level events and absent for tour-level ones")
	ErrBadPeriod        = errors.New("'to' must be after 'from'")

	ErrEventKeyRequired    = errors.New("event_key is required")
	ErrVersionRequired     = errors.New("tour_version_id is required")
	ErrTourRequired        = errors.New("tour_id is required")
	ErrVersionForeignApp   = errors.New("tour version belongs to another application")
	ErrTourVersionMismatch = errors.New("tour version does not belong to the specified tour")
	ErrHintNotInVersion    = errors.New("hint does not belong to the specified tour version")
	ErrCorruptedData       = errors.New("corrupted JSONB data in database")
)

// ---------------------------- //
// ----- Validation Error ----- //
// ---------------------------- //

type (
	ErrorDetail struct {
		Path    string
		Message string
	}

	ValidationError struct {
		Err     error
		Details []ErrorDetail
	}
)

func (ve *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %v (%d problems)", ve.Err, len(ve.Details))
}

func (ve *ValidationError) Message() string { return capitalise(ve.Err.Error()) }

func (ve *ValidationError) Unwrap() error { return ve.Err }

func (ve *ValidationError) StatusCode() int {
	switch ve.Err {
	case ErrTourNotPublishable:
		return http.StatusUnprocessableEntity
	case ErrDraftInvalid:
		return http.StatusBadRequest
	default:
		return http.StatusBadRequest
	}

}

func (ve *ValidationError) ToHTTPError() *models.HTTPError {
	problems := make([]string, 0, len(ve.Details))
	for _, d := range ve.Details {
		problems = append(problems, fmt.Sprintf("%s: %s", d.Path, d.Message))
	}
	return &models.HTTPError{
		Error:   fmt.Sprintf("%v: %s", ve.Err, strings.Join(problems, "; ")),
		Message: ve.Message(),
	}
}

// --------------------- //
// ----- Helpers ------- //
// --------------------- //

func logicErr(err error, stage string, code int) error {
	return &LogicError{Err: err, Stage: stage, Code: code}
}

func NotFound(sentinel error, ref string) error {
	return &RepositoryError{
		Err:       sentinel,
		EntityRef: ref,
		Operation: Retrieve,
		Reason:    NoRecord,
	}
}

func IsNotFound(err error) bool {
	var repoErr *RepositoryError
	if errors.As(err, &repoErr) {
		return repoErr.Reason == NoRecord
	}
	return false
}
