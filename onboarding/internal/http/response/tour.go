package response

import (
	"time"

	"github.com/Vanady39/cluer/onboarding/internal/domains"
	"github.com/google/uuid"
)

// These are the shapes the API promises. They mirror the domain today, and
// that is the point: the mirror is what lets the domain change tomorrow
// without breaking a client. Renaming a field here is a breaking API change;
// renaming one in domains is not.

type TriggerConfig struct {
	ScrollDepth     *int    `json:"scroll_depth,omitempty" example:"50"`
	InactivitySecs  *int    `json:"inactivity_secs,omitempty" example:"30"`
	ElementSelector *string `json:"element_selector,omitempty"`
	DelayMs         *int    `json:"delay_ms,omitempty" example:"1500"`
}

type AudienceRule struct {
	Type      string `json:"type" example:"page_visited" enums:"event_performed,page_visited,user_property"`
	Key       string `json:"key" example:"is_premium"`
	Operator  string `json:"operator" example:"eq" enums:"exists,not_exists,eq,neq,gt,gte,lt,lte,contains,not_contains,starts_with,ends_with"`
	Value     any    `json:"value,omitempty"`
	Timeframe string `json:"timeframe,omitempty" example:"7d" enums:"1h,24h,7d,30d,all_time"`
}

type Audience struct {
	ShowOnce bool           `json:"show_once" example:"true"`
	MaxShows int            `json:"max_shows" example:"3"`
	OnlyNew  bool           `json:"only_new" example:"false"`
	Rules    []AudienceRule `json:"rules,omitempty"`
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

	VersionId     uuid.UUID      `json:"version_id,omitzero" format:"uuid"`
	Version       int            `json:"version,omitempty" example:"7"`
	Status        string         `json:"status,omitempty" example:"draft" enums:"draft,published,archived"`
	TriggerType   string         `json:"trigger_type,omitempty" example:"on_load" enums:"on_load,delay,exit_intent,manual,scroll_depth,inactivity,element_visible"`
	TriggerConfig *TriggerConfig `json:"trigger_config,omitempty"`

	TargetPath string    `json:"target_path,omitempty" example:"/dashboard"`
	Audience   *Audience `json:"audience,omitempty"`
	Hints      []Hint    `json:"hints,omitempty"`
}

type TourVersion struct {
	Id            uuid.UUID      `json:"id" format:"uuid"`
	TourId        uuid.UUID      `json:"tour_id" format:"uuid"`
	Version       int            `json:"version" example:"7"`
	Status        string         `json:"status" example:"published" enums:"draft,published,archived"`
	TriggerType   string         `json:"trigger_type" example:"on_load" enums:"on_load,delay,exit_intent,manual,scroll_depth,inactivity,element_visible"`
	TriggerConfig *TriggerConfig `json:"trigger_config,omitempty"`
	TargetPath    string         `json:"target_path" example:"/dashboard"`
	Audience      Audience       `json:"audience"`
	CreatedBy     string         `json:"created_by,omitempty"`
	CreatedAt     time.Time      `json:"created_at" format:"date-time"`
	PublishedAt   *time.Time     `json:"published_at,omitempty" format:"date-time"`
	ArchivedAt    *time.Time     `json:"archived_at,omitempty" format:"date-time"`
	Hints         []Hint         `json:"hints,omitempty"`
}

type TourCard struct {
	Tour      *Tour        `json:"tour"`
	Published *TourVersion `json:"published"`
	Draft     *TourVersion `json:"draft"`
}

func NewTriggerConfig(cfg *domains.TriggerConfig) *TriggerConfig {
	if cfg == nil {
		return nil
	}
	return &TriggerConfig{
		ScrollDepth:     cfg.ScrollDepth,
		InactivitySecs:  cfg.InactivitySecs,
		ElementSelector: cfg.ElementSelector,
		DelayMs:         cfg.DelayMs,
	}
}

// newAudienceRules keeps a nil slice nil: "no rules" and "an empty rule list"
// serialise the same way today, but the distinction is free to preserve and
// costly to reintroduce.
func newAudienceRules(rules []domains.AudienceRule) []AudienceRule {
	if rules == nil {
		return nil
	}
	out := make([]AudienceRule, 0, len(rules))
	for _, rule := range rules {
		out = append(out, AudienceRule{
			Type:      rule.Type,
			Key:       rule.Key,
			Operator:  rule.Operator,
			Value:     rule.Value,
			Timeframe: rule.Timeframe,
		})
	}
	return out
}

func newAudienceValue(a domains.Audience) Audience {
	return Audience{
		ShowOnce: a.ShowOnce,
		MaxShows: a.MaxShows,
		OnlyNew:  a.OnlyNew,
		Rules:    newAudienceRules(a.Rules),
	}
}

func NewAudience(a *domains.Audience) *Audience {
	if a == nil {
		return nil
	}
	out := newAudienceValue(*a)
	return &out
}

func NewTour(t *domains.Tour) *Tour {
	if t == nil {
		return nil
	}
	return &Tour{
		Id:            t.Id,
		AppId:         t.AppId,
		Title:         t.Title,
		Description:   t.Description,
		Enabled:       t.Enabled,
		Priority:      t.Priority,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
		VersionId:     t.VersionId,
		Version:       t.Version,
		Status:        string(t.Status),
		TriggerType:   string(t.TriggerType),
		TriggerConfig: NewTriggerConfig(t.TriggerConfig),
		TargetPath:    t.TargetPath,
		Audience:      NewAudience(t.Audience),
		Hints:         NewHints(t.Hints),
	}
}

func NewTours(tours []*domains.Tour) []*Tour {
	if tours == nil {
		return nil
	}
	out := make([]*Tour, 0, len(tours))
	for _, t := range tours {
		out = append(out, NewTour(t))
	}
	return out
}

func NewTourVersion(v *domains.TourVersion) *TourVersion {
	if v == nil {
		return nil
	}
	return &TourVersion{
		Id:            v.Id,
		TourId:        v.TourId,
		Version:       v.Version,
		Status:        string(v.Status),
		TriggerType:   string(v.TriggerType),
		TriggerConfig: NewTriggerConfig(v.TriggerConfig),
		TargetPath:    v.TargetPath,
		Audience:      newAudienceValue(v.Audience),
		CreatedBy:     v.CreatedBy,
		CreatedAt:     v.CreatedAt,
		PublishedAt:   v.PublishedAt,
		ArchivedAt:    v.ArchivedAt,
		Hints:         NewHints(v.Hints),
	}
}

func NewTourVersions(versions []*domains.TourVersion) []*TourVersion {
	if versions == nil {
		return nil
	}
	out := make([]*TourVersion, 0, len(versions))
	for _, v := range versions {
		out = append(out, NewTourVersion(v))
	}
	return out
}

func NewTourCard(c *domains.TourCard) *TourCard {
	if c == nil {
		return nil
	}
	return &TourCard{
		Tour:      NewTour(c.Tour),
		Published: NewTourVersion(c.Published),
		Draft:     NewTourVersion(c.Draft),
	}
}
