package domains

import (
	"time"

	"github.com/google/uuid"
)

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

type Audience struct {
	ShowOnce bool `json:"show_once" example:"true"`
	MaxShows int  `json:"max_shows" example:"3"`
	OnlyNew  bool `json:"only_new" example:"false"`
}

type Tour struct {
	Id          uuid.UUID `json:"id" format:"uuid" example:"01914f6a-8c9b-7a0b-9b0b-8c9b7a0b9b0b"`
	AppId       uuid.UUID `json:"app_id" format:"uuid"`
	Title       string    `json:"title" example:"Добро пожаловать в систему"`
	Description string    `json:"description" example:"Тур по основным функциям дашборда"`
	Enabled     bool      `json:"enabled" example:"true"`
	Priority    int       `json:"priority" example:"1"`
	CreatedAt   time.Time `json:"created_at" format:"date-time"`
	UpdatedAt   time.Time `json:"updated_at" format:"date-time"`

	VersionId   uuid.UUID   `json:"version_id,omitzero" format:"uuid"`
	Version     int         `json:"version,omitempty" example:"7"`
	Status      TourStatus  `json:"status,omitempty" example:"draft"`
	TriggerType TriggerType `json:"trigger_type,omitempty" example:"on_load"`
	TargetPath  string      `json:"target_path,omitempty" example:"/dashboard"`
	Audience    *Audience   `json:"audience,omitempty"`
	Hints       []Hint      `json:"hints,omitempty"`
}

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

type TourCard struct {
	Tour      *Tour        `json:"tour"`
	Published *TourVersion `json:"published"`
	Draft     *TourVersion `json:"draft"`
}

func (v *TourVersion) AsTour(t *Tour) *Tour {
	merged := *t
	merged.VersionId = v.Id
	merged.Version = v.Version
	merged.Status = v.Status
	merged.TriggerType = v.TriggerType
	merged.TargetPath = v.TargetPath
	audience := v.Audience
	merged.Audience = &audience
	merged.Hints = v.Hints
	if merged.Hints == nil {
		merged.Hints = []Hint{}
	}
	return &merged
}
