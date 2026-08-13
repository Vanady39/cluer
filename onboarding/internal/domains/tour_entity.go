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
	TriggerOnLoad         TriggerType = "on_load"
	TriggerDelay          TriggerType = "delay"
	TriggerExitIntent     TriggerType = "exit_intent"
	TriggerManual         TriggerType = "manual"
	TriggerScrollDepth    TriggerType = "scroll_depth"
	TriggerInactivity     TriggerType = "inactivity"
	TriggerElementVisible TriggerType = "element_visible"
)

type TriggerConfig struct {
	ScrollDepth     *int
	InactivitySecs  *int
	ElementSelector *string
	DelayMs         *int
}

func (t TriggerType) Valid() bool {
	switch t {
	case TriggerOnLoad, TriggerDelay, TriggerExitIntent, TriggerManual,
		TriggerScrollDepth, TriggerInactivity, TriggerElementVisible:
		return true
	}
	return false
}

type AudienceRule struct {
	Type      string
	Key       string
	Operator  string
	Value     any
	Timeframe string
}
type Audience struct {
	ShowOnce bool
	MaxShows int
	OnlyNew  bool
	Rules    []AudienceRule
}

type Tour struct {
	Id          uuid.UUID
	AppId       uuid.UUID
	Title       string
	Description string
	Enabled     bool
	Priority    int
	CreatedAt   time.Time
	UpdatedAt   time.Time

	VersionId     uuid.UUID
	Version       int
	Status        TourStatus
	TriggerType   TriggerType
	TriggerConfig *TriggerConfig

	TargetPath string
	Audience   *Audience
	Hints      []Hint
}

type TourVersion struct {
	Id            uuid.UUID
	TourId        uuid.UUID
	Version       int
	Status        TourStatus
	TriggerType   TriggerType
	TriggerConfig *TriggerConfig
	TargetPath    string
	Audience      Audience
	CreatedBy     string
	CreatedAt     time.Time
	PublishedAt   *time.Time
	ArchivedAt    *time.Time
	Hints         []Hint
}

type TourCard struct {
	Tour      *Tour
	Published *TourVersion
	Draft     *TourVersion
}

func (v *TourVersion) AsTour(t *Tour) *Tour {
	merged := *t
	merged.VersionId = v.Id
	merged.Version = v.Version
	merged.Status = v.Status
	merged.TriggerType = v.TriggerType
	merged.TriggerConfig = v.TriggerConfig
	merged.TargetPath = v.TargetPath
	audience := v.Audience
	merged.Audience = &audience
	merged.Hints = v.Hints
	if merged.Hints == nil {
		merged.Hints = []Hint{}
	}
	return &merged
}
