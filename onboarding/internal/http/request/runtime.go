package request

import (
	"time"

	"github.com/Vanady39/cluer/onboarding/internal/domains"
	"github.com/google/uuid"
)

type ResolveRequest struct {
	Url       string         `json:"url"`
	SubjectId string         `json:"subject_id" binding:"required"`
	SessionId string         `json:"session_id"`
	Props     map[string]any `json:"props"`
}

type EventItem struct {
	EventKey      string         `json:"event_key" binding:"required"`
	Type          string         `json:"type" binding:"required" enums:"tour_started,hint_shown,hint_completed,hint_skipped,selector_missing,tour_completed,tour_dismissed,goal_reached,custom"`
	TourId        uuid.UUID      `json:"tour_id"`
	TourVersionId uuid.UUID      `json:"tour_version_id"`
	HintId        *uuid.UUID     `json:"hint_id"`
	OccurredAt    *time.Time     `json:"occurred_at"`
	Payload       map[string]any `json:"payload"`
}

type EventBatchRequest struct {
	SubjectId string      `json:"subject_id" binding:"required"`
	SessionId string      `json:"session_id" binding:"required"`
	Events    []EventItem `json:"events" binding:"required"`
}

type CreateAppRequest struct {
	Name           string   `json:"name" binding:"required"`
	AllowedOrigins []string `json:"allowed_origins"`
}

// ToDomain fills only what the client is allowed to state. AppId, SubjectId and
// SessionId are taken from the request context and the batch envelope, never
// from the item, so a caller cannot post events on behalf of another app.
func (e *EventItem) ToDomain() domains.Event {
	event := domains.Event{
		EventKey:      e.EventKey,
		TourId:        e.TourId,
		TourVersionId: e.TourVersionId,
		HintId:        e.HintId,
		Type:          domains.EventType(e.Type),
		Payload:       e.Payload,
	}
	if e.OccurredAt != nil {
		event.OccurredAt = *e.OccurredAt
	}
	return event
}

func (b *EventBatchRequest) EventsToDomain() []domains.Event {
	events := make([]domains.Event, 0, len(b.Events))
	for i := range b.Events {
		events = append(events, b.Events[i].ToDomain())
	}
	return events
}

func (r *ResolveRequest) ToDomain(appId uuid.UUID) domains.ResolveRequest {
	return domains.ResolveRequest{
		AppId:     appId,
		Url:       r.Url,
		SubjectId: r.SubjectId,
		SessionId: r.SessionId,
		Props:     r.Props,
	}
}

func (a *CreateAppRequest) ToDomain() *domains.App {
	return &domains.App{
		Name:           a.Name,
		AllowedOrigins: a.AllowedOrigins,
	}
}
