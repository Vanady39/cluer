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
	Id          uuid.UUID   `json:"id" db:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title       string      `json:"title" db:"title" example:"Добро пожаловать в систему"`
	Description string      `json:"description" db:"description" example:"Тур по основным функциям дашборда"`
	Status      TourStatus  `json:"status" db:"status" example:"draft"`
	TriggerType TriggerType `json:"trigger_type" db:"trigger_type" example:"on_load"`
	TargetPath  string      `json:"target_path" db:"target_path" example:"/dashboard"`
	Priority    int         `json:"priority" db:"priority" example:"1"`
	Audience    Audience    `json:"audience" db:"audience"`
	Hints       []Hint      `json:"hints" db:"hints"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}
