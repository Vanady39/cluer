package domains

import (
	"time"

	"github.com/google/uuid"
)

type App struct {
	Id             uuid.UUID
	Name           string
	PublicKey      string
	AllowedOrigins []string
	CreatedAt      time.Time
	ArchivedAt     *time.Time
}

type ProgressStatus string

const (
	ProgressInProgress ProgressStatus = "in_progress"
	ProgressCompleted  ProgressStatus = "completed"
	ProgressDismissed  ProgressStatus = "dismissed"
)

type Progress struct {
	AppId         uuid.UUID
	SubjectId     string
	TourId        uuid.UUID
	TourVersionId uuid.UUID
	CurrentHintId *uuid.UUID
	Status        ProgressStatus
	ShowsCount    int
	StartedAt     time.Time
	UpdatedAt     time.Time
	FinishedAt    *time.Time
}

type EventType string

const (
	EventTourStarted     EventType = "tour_started"
	EventHintShown       EventType = "hint_shown"
	EventHintCompleted   EventType = "hint_completed"
	EventHintSkipped     EventType = "hint_skipped"
	EventSelectorMissing EventType = "selector_missing"
	EventTourCompleted   EventType = "tour_completed"
	EventTourDismissed   EventType = "tour_dismissed"
	EventGoalReached     EventType = "goal_reached"
	EventCustom          EventType = "custom"
)

func (t EventType) Valid() bool {
	switch t {
	case EventTourStarted, EventHintShown, EventHintCompleted, EventHintSkipped,
		EventSelectorMissing, EventTourCompleted, EventTourDismissed, EventGoalReached, EventCustom:
		return true
	}
	return false
}

func (t EventType) TourLevel() bool {
	switch t {
	case EventTourStarted, EventTourCompleted, EventTourDismissed, EventGoalReached, EventCustom:
		return true
	}
	return false
}

type Event struct {
	Id            int64
	AppId         uuid.UUID
	EventKey      string
	TourId        uuid.UUID
	TourVersionId uuid.UUID
	HintId        *uuid.UUID
	SessionId     string
	SubjectId     string
	Type          EventType
	OccurredAt    time.Time
	ReceivedAt    time.Time
	Payload       map[string]any
}

type ResolveRequest struct {
	AppId     uuid.UUID
	Url       string
	SubjectId string
	SessionId string
	Props     map[string]any
}

type ResolveResult struct {
	Tour          *Tour
	TourVersionId uuid.UUID
	Version       int
	CurrentHintId *uuid.UUID
}

type EventBatchResult struct {
	Accepted   int
	Duplicates int
	Rejected   int
	Errors     []string
}

func (r *EventBatchResult) reject(err error) {
	r.Rejected++
	r.Errors = append(r.Errors, err.Error())
}

type FunnelStep struct {
	Step            int
	HintId          uuid.UUID
	Title           string
	Shown           int
	Completed       int
	Skipped         int
	SelectorMissing int
	Dropoff         float64
}

type AnalyticsTotals struct {
	Started        int
	Completed      int
	Dismissed      int
	GoalReached    int
	CompletionRate float64
	GoalRate       float64
}

type Analytics struct {
	TourId          uuid.UUID
	TourVersionId   uuid.UUID
	Version         int
	From            time.Time
	To              time.Time
	Totals          AnalyticsTotals
	Funnel          []FunnelStep
	BrokenSelectors []uuid.UUID
}
