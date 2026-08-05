package onboarding

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
	ShowOnce bool `json:"show_once"`
	MaxShows int  `json:"max_shows"`
	OnlyNew  bool `json:"only_new"`
}

type Tour struct {
	Id          uuid.UUID   `json:"id" db:"id"`
	Title       string      `json:"title" db:"title"`
	Description string      `json:"description" db:"description"`
	Status      TourStatus  `json:"status" db:"status"`
	TriggerType TriggerType `json:"trigger_type" db:"trigger_type"`
	TargetPath  string      `json:"target_path" db:"target_path"`
	Priority    int         `json:"priority" db:"priority"`
	Audience    Audience    `json:"audience" db:"audience"`
	Hints       []Hint      `json:"hints" db:"hints"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}
