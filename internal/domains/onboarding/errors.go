package onboarding

import "errors"

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
