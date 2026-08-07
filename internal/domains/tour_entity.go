package domains

import (
	"time"

	"github.com/google/uuid"
)

// TourStatus is the status of a *version*, not of the tour. A tour has no
// status of its own: what is live is derived from which of its versions is
// published, and the partial unique index in the schema guarantees there is at
// most one such version.
type TourStatus string

const (
	TourDraft     TourStatus = "draft"
	TourPublished TourStatus = "published"
	TourArchived  TourStatus = "archived"
)

func (s TourStatus) Valid() bool {
	switch s {
	case TourDraft, TourPublished, TourArchived:
		return true
	}
	return false
}

type TriggerType string

const (
	TriggerOnLoad     TriggerType = "on_load"
	TriggerDelay      TriggerType = "delay"
	TriggerExitIntent TriggerType = "exit_intent"
	TriggerManual     TriggerType = "manual"
)

func (t TriggerType) Valid() bool {
	switch t {
	case TriggerOnLoad, TriggerDelay, TriggerExitIntent, TriggerManual:
		return true
	}
	return false
}

// Audience targets the tour. Evaluated on the backend, never in the browser:
// targeting logic handed to the client is targeting logic that can be edited by
// the client.
type Audience struct {
	// ShowOnce suppresses the tour for subjects who already completed or dismissed it.
	ShowOnce bool `json:"show_once" example:"true"`
	// MaxShows caps how many times the tour may be resolved for one subject. 0 = unlimited.
	MaxShows int `json:"max_shows" example:"3"`
	// OnlyNew restricts the tour to subjects the host app marks as new.
	OnlyNew bool `json:"only_new" example:"false"`
}

// Tour is the wire shape: the stable entity and the fields of whichever version
// is being represented, flattened into one object. Two tables underneath, one
// document on the wire — the API contract predates the versioning and there is
// no reason to make every consumer learn about the split.
type Tour struct {
	Id          uuid.UUID `json:"id" format:"uuid" example:"01914f6a-8c9b-7a0b-9b0b-8c9b7a0b9b0b"`
	AppId       uuid.UUID `json:"app_id" format:"uuid"`
	Title       string    `json:"title" example:"Добро пожаловать в систему"`
	Description string    `json:"description" example:"Тур по основным функциям дашборда"`
	Enabled     bool      `json:"enabled" example:"true"`
	Priority    int       `json:"priority" example:"1"`
	CreatedAt   time.Time `json:"created_at" format:"date-time"`
	UpdatedAt   time.Time `json:"updated_at" format:"date-time"`

	// --- fields owned by the version ---
	//
	// Omitted when the payload carries no version, which happens on the tour
	// nested inside a TourCard and on the result of a metadata patch. Emitting
	// them as empty strings would read as "this tour has no target path" rather
	// than "this response is not about a version".
	VersionId   uuid.UUID   `json:"version_id,omitzero" format:"uuid"`
	Version     int         `json:"version,omitempty" example:"7"`
	Status      TourStatus  `json:"status,omitempty" example:"draft"`
	TriggerType TriggerType `json:"trigger_type,omitempty" example:"on_load"`
	TargetPath  string      `json:"target_path,omitempty" example:"/dashboard"`
	Audience    Audience    `json:"audience,omitzero"`
	Hints       []Hint      `json:"hints,omitempty"`
}

// TourVersion is an immutable edition of a tour. Every analytics event carries
// its id, which is what makes the reference an exact snapshot of what the user
// saw on screen.
type TourVersion struct {
	Id          uuid.UUID   `json:"id" format:"uuid"`
	TourId      uuid.UUID   `json:"tour_id" format:"uuid"`
	Version     int         `json:"version" example:"7"`
	Status      TourStatus  `json:"status" example:"published"`
	TriggerType TriggerType `json:"trigger_type" example:"on_load"`
	TargetPath  string      `json:"target_path" example:"/dashboard"`
	Audience    Audience    `json:"audience"`
	CreatedBy   string      `json:"created_by,omitempty"`
	CreatedAt   time.Time   `json:"created_at" format:"date-time"`
	PublishedAt *time.Time  `json:"published_at,omitempty" format:"date-time"`
	ArchivedAt  *time.Time  `json:"archived_at,omitempty" format:"date-time"`
	Hints       []Hint      `json:"hints,omitempty"`
}

// TourCard is what the admin editor opens: the tour plus both versions that can
// exist at once. Returning them together saves the editor from three round
// trips to render two columns.
type TourCard struct {
	Tour      *Tour        `json:"tour"`
	Published *TourVersion `json:"published"`
	Draft     *TourVersion `json:"draft"`
}

// AsTour merges a version back into the flat wire shape.
func (v *TourVersion) AsTour(t *Tour) *Tour {
	merged := *t
	merged.VersionId = v.Id
	merged.Version = v.Version
	merged.Status = v.Status
	merged.TriggerType = v.TriggerType
	merged.TargetPath = v.TargetPath
	merged.Audience = v.Audience
	merged.Hints = v.Hints
	if merged.Hints == nil {
		merged.Hints = []Hint{}
	}
	return &merged
}
