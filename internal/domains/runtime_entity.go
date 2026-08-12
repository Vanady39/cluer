package domains

import (
	"time"

	"github.com/google/uuid"
)

type App struct {
	Id             uuid.UUID  `json:"id" format:"uuid"`
	Name           string     `json:"name" example:"Demo classifieds"`
	PublicKey      string     `json:"public_key" example:"pk_demo_7f3a91c4b2e85d60"`
	AllowedOrigins []string   `json:"allowed_origins" example:"http://localhost:3000"`
	CreatedAt      time.Time  `json:"created_at" format:"date-time"`
	ArchivedAt     *time.Time `json:"archived_at,omitempty" format:"date-time"`
}

type ProgressStatus string

const (
	ProgressInProgress ProgressStatus = "in_progress"
	ProgressCompleted  ProgressStatus = "completed"
	ProgressDismissed  ProgressStatus = "dismissed"
)

type Progress struct {
	AppId         uuid.UUID      `json:"app_id" format:"uuid"`
	SubjectId     string         `json:"subject_id" example:"anon_8f3d2c"`
	TourId        uuid.UUID      `json:"tour_id" format:"uuid"`
	TourVersionId uuid.UUID      `json:"tour_version_id" format:"uuid"`
	CurrentHintId *uuid.UUID     `json:"current_hint_id,omitempty" format:"uuid"`
	Status        ProgressStatus `json:"status" example:"in_progress"`
	ShowsCount    int            `json:"shows_count" example:"1"`
	StartedAt     time.Time      `json:"started_at" format:"date-time"`
	UpdatedAt     time.Time      `json:"updated_at" format:"date-time"`
	FinishedAt    *time.Time     `json:"finished_at,omitempty" format:"date-time"`
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
	Id            int64          `json:"id"`
	AppId         uuid.UUID      `json:"app_id" format:"uuid"`
	EventKey      string         `json:"event_key" example:"b1e7a0c2-19f4-4a3d-9c11-2f8e5a6d0b73"`
	TourId        uuid.UUID      `json:"tour_id" format:"uuid"`
	TourVersionId uuid.UUID      `json:"tour_version_id" format:"uuid"`
	HintId        *uuid.UUID     `json:"hint_id,omitempty" format:"uuid"`
	SessionId     string         `json:"session_id" example:"sess_2b91f0"`
	SubjectId     string         `json:"subject_id" example:"anon_8f3d2c"`
	Type          EventType      `json:"type" example:"hint_shown"`
	OccurredAt    time.Time      `json:"occurred_at" format:"date-time"`
	ReceivedAt    time.Time      `json:"received_at" format:"date-time"`
	Payload       map[string]any `json:"payload"`
}

type ResolveRequest struct {
	AppId     uuid.UUID
	Url       string
	SubjectId string
	SessionId string
	Props     map[string]any
}

type ResolveResult struct {
	Tour          *Tour      `json:"tour"`
	TourVersionId uuid.UUID  `json:"tour_version_id" format:"uuid"`
	Version       int        `json:"version" example:"7"`
	CurrentHintId *uuid.UUID `json:"current_hint_id,omitempty" format:"uuid"`
}

type EventBatchResult struct {
	Accepted   int      `json:"accepted" example:"4"`
	Duplicates int      `json:"duplicates" example:"1"`
	Rejected   int      `json:"rejected" example:"0"`
	Errors     []string `json:"errors,omitempty"`
}

func (r *EventBatchResult) reject(err error) {
	r.Rejected++
	r.Errors = append(r.Errors, err.Error())
}

type FunnelStep struct {
	Step            int       `json:"step" example:"1"`
	HintId          uuid.UUID `json:"hint_id" format:"uuid"`
	Title           string    `json:"title" example:"Начните здесь"`
	Shown           int       `json:"shown" example:"1240"`
	Completed       int       `json:"completed" example:"1105"`
	Skipped         int       `json:"skipped" example:"0"`
	SelectorMissing int       `json:"selector_missing" example:"0"`
	Dropoff         float64   `json:"dropoff" example:"0.109"`
}

type AnalyticsTotals struct {
	Started        int     `json:"started" example:"1240"`
	Completed      int     `json:"completed" example:"612"`
	Dismissed      int     `json:"dismissed" example:"388"`
	GoalReached    int     `json:"goal_reached" example:"501"`
	CompletionRate float64 `json:"completion_rate" example:"0.494"`
	GoalRate       float64 `json:"goal_rate" example:"0.404"`
}

type Analytics struct {
	TourId          uuid.UUID       `json:"tour_id" format:"uuid"`
	TourVersionId   uuid.UUID       `json:"tour_version_id" format:"uuid"`
	Version         int             `json:"version" example:"7"`
	From            time.Time       `json:"from" format:"date-time"`
	To              time.Time       `json:"to" format:"date-time"`
	Totals          AnalyticsTotals `json:"totals"`
	Funnel          []FunnelStep    `json:"funnel"`
	BrokenSelectors []uuid.UUID     `json:"broken_selectors"`
}
