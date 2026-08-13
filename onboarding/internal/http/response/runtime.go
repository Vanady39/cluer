package response

import (
	"time"

	"github.com/Vanady39/cluer/onboarding/internal/domains"
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

// Health is the only response here with no domain counterpart: liveness is a
// property of the process, not of the business.
type Health struct {
	Status string `json:"status" example:"ok"`
	DB     string `json:"db" example:"ok"`
}

func NewApp(a *domains.App) *App {
	if a == nil {
		return nil
	}
	return &App{
		Id:             a.Id,
		Name:           a.Name,
		PublicKey:      a.PublicKey,
		AllowedOrigins: a.AllowedOrigins,
		CreatedAt:      a.CreatedAt,
		ArchivedAt:     a.ArchivedAt,
	}
}

func NewApps(apps []*domains.App) []*App {
	if apps == nil {
		return nil
	}
	out := make([]*App, 0, len(apps))
	for _, a := range apps {
		out = append(out, NewApp(a))
	}
	return out
}

func NewResolveResult(r *domains.ResolveResult) *ResolveResult {
	if r == nil {
		return nil
	}
	return &ResolveResult{
		Tour:          NewTour(r.Tour),
		TourVersionId: r.TourVersionId,
		Version:       r.Version,
		CurrentHintId: r.CurrentHintId,
	}
}

func NewEventBatchResult(r *domains.EventBatchResult) *EventBatchResult {
	if r == nil {
		return nil
	}
	return &EventBatchResult{
		Accepted:   r.Accepted,
		Duplicates: r.Duplicates,
		Rejected:   r.Rejected,
		Errors:     r.Errors,
	}
}

func newFunnel(steps []domains.FunnelStep) []FunnelStep {
	if steps == nil {
		return nil
	}
	out := make([]FunnelStep, 0, len(steps))
	for _, s := range steps {
		out = append(out, FunnelStep{
			Step:            s.Step,
			HintId:          s.HintId,
			Title:           s.Title,
			Shown:           s.Shown,
			Completed:       s.Completed,
			Skipped:         s.Skipped,
			SelectorMissing: s.SelectorMissing,
			Dropoff:         s.Dropoff,
		})
	}
	return out
}

func NewAnalytics(a *domains.Analytics) *Analytics {
	if a == nil {
		return nil
	}
	return &Analytics{
		TourId:        a.TourId,
		TourVersionId: a.TourVersionId,
		Version:       a.Version,
		From:          a.From,
		To:            a.To,
		Totals: AnalyticsTotals{
			Started:        a.Totals.Started,
			Completed:      a.Totals.Completed,
			Dismissed:      a.Totals.Dismissed,
			GoalReached:    a.Totals.GoalReached,
			CompletionRate: a.Totals.CompletionRate,
			GoalRate:       a.Totals.GoalRate,
		},
		Funnel:          newFunnel(a.Funnel),
		BrokenSelectors: a.BrokenSelectors,
	}
}
