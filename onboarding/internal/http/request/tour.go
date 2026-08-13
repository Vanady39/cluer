// Package request holds what the API accepts. Bodies are bound here and only
// then translated into domain values, so a client cannot reach a domain field
// the API never promised — and the domain can gain or rename fields without
// silently widening the accepted payload.
package request

import (
	"time"

	"github.com/Vanady39/cluer/onboarding/internal/domains"
	"github.com/google/uuid"
)

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
	Id          uuid.UUID `json:"id" format:"uuid"`
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

type TourMetaPatch struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Enabled     *bool   `json:"enabled"`
	Priority    *int    `json:"priority"`
}

type DraftPatch struct {
	TriggerType   *string        `json:"trigger_type"`
	TriggerConfig *TriggerConfig `json:"trigger_config"`
	TargetPath    *string        `json:"target_path"`
	Audience      *Audience      `json:"audience"`
}

type RollbackRequest struct {
	ToVersionId uuid.UUID `json:"to_version_id" binding:"required"`
}

func (c *TriggerConfig) ToDomain() *domains.TriggerConfig {
	if c == nil {
		return nil
	}
	return &domains.TriggerConfig{
		ScrollDepth:     c.ScrollDepth,
		InactivitySecs:  c.InactivitySecs,
		ElementSelector: c.ElementSelector,
		DelayMs:         c.DelayMs,
	}
}

func audienceRulesToDomain(rules []AudienceRule) []domains.AudienceRule {
	if rules == nil {
		return nil
	}
	out := make([]domains.AudienceRule, 0, len(rules))
	for _, rule := range rules {
		out = append(out, domains.AudienceRule{
			Type:      rule.Type,
			Key:       rule.Key,
			Operator:  rule.Operator,
			Value:     rule.Value,
			Timeframe: rule.Timeframe,
		})
	}
	return out
}

func (a *Audience) ToDomain() *domains.Audience {
	if a == nil {
		return nil
	}
	return &domains.Audience{
		ShowOnce: a.ShowOnce,
		MaxShows: a.MaxShows,
		OnlyNew:  a.OnlyNew,
		Rules:    audienceRulesToDomain(a.Rules),
	}
}

func (t *Tour) ToDomain() *domains.Tour {
	if t == nil {
		return nil
	}
	return &domains.Tour{
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
		Status:        domains.TourStatus(t.Status),
		TriggerType:   domains.TriggerType(t.TriggerType),
		TriggerConfig: t.TriggerConfig.ToDomain(),
		TargetPath:    t.TargetPath,
		Audience:      t.Audience.ToDomain(),
		Hints:         HintsToDomain(t.Hints),
	}
}

func (p *TourMetaPatch) ToDomain() domains.TourMetaPatch {
	return domains.TourMetaPatch{
		Title:       p.Title,
		Description: p.Description,
		Enabled:     p.Enabled,
		Priority:    p.Priority,
	}
}

func (p *DraftPatch) ToDomain() domains.DraftPatch {
	out := domains.DraftPatch{
		TargetPath:    p.TargetPath,
		TriggerConfig: p.TriggerConfig.ToDomain(),
		Audience:      p.Audience.ToDomain(),
	}
	if p.TriggerType != nil {
		triggerType := domains.TriggerType(*p.TriggerType)
		out.TriggerType = &triggerType
	}
	return out
}
