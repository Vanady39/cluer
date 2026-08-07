package domains

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const maxEventsPerBatch = 50

type (
	RuntimeDomain struct {
		tours   TourRepositoryInterface
		hints   HintRepositoryInterface
		runtime RuntimeRepositoryInterface
	}

	RuntimeDomainInterface interface {
		AppByKey(ctx context.Context, key string) (*App, error)
		CreateApp(ctx context.Context, a *App) (*App, error)
		ListApps(ctx context.Context) ([]*App, error)

		Resolve(ctx context.Context, req ResolveRequest) (*ResolveResult, error)
		Ingest(ctx context.Context, appId uuid.UUID, subjectId, sessionId string, events []Event) (*EventBatchResult, error)
		Analytics(ctx context.Context, tourId, versionId uuid.UUID, from, to time.Time) (*Analytics, error)
		Ping(ctx context.Context) error
	}

	RuntimeRepositoryInterface interface {
		CreateApp(ctx context.Context, a *App) error
		AppByKey(ctx context.Context, key string) (*App, error)
		ListApps(ctx context.Context) ([]*App, error)

		SubjectProgress(ctx context.Context, appId uuid.UUID, subjectId string) (map[uuid.UUID]*Progress, error)
		UpsertProgress(ctx context.Context, p *Progress) error

		// InsertEvents is idempotent on (app_id, event_key) and applies the
		// progress side effects in the same transaction.
		InsertEvents(ctx context.Context, events []Event) (accepted, duplicates int, err error)
		Funnel(ctx context.Context, versionId uuid.UUID, from, to time.Time) ([]FunnelStep, error)
		Totals(ctx context.Context, versionId uuid.UUID, from, to time.Time) (AnalyticsTotals, error)
		Ping(ctx context.Context) error
	}
)

func NewRuntimeDomain(
	tours TourRepositoryInterface,
	hints HintRepositoryInterface,
	runtime RuntimeRepositoryInterface,
) *RuntimeDomain {
	return &RuntimeDomain{tours: tours, hints: hints, runtime: runtime}
}

func (rd *RuntimeDomain) AppByKey(ctx context.Context, key string) (*App, error) {
	return rd.runtime.AppByKey(ctx, key)
}

func (rd *RuntimeDomain) CreateApp(ctx context.Context, a *App) (*App, error) {
	if a.Name == "" {
		return nil, logicErr(ErrTitleRequired, "app validation", http.StatusBadRequest)
	}
	if a.PublicKey == "" {
		a.PublicKey = "pk_" + uuid.NewString()
	}
	if a.AllowedOrigins == nil {
		a.AllowedOrigins = []string{}
	}
	if err := rd.runtime.CreateApp(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (rd *RuntimeDomain) ListApps(ctx context.Context) ([]*App, error) {
	return rd.runtime.ListApps(ctx)
}

func (rd *RuntimeDomain) Ping(ctx context.Context) error {
	return rd.runtime.Ping(ctx)
}

// Resolve answers "what do we show this user, here, right now" in one request.
//
// The whole tour goes back in that one response on purpose. Asking the backend
// for each hint would put a 200-400ms round trip on every click, which reads as
// lag, and a dropped connection would tear the onboarding in half. Sending it
// whole means a dropped connection costs analytics, not the onboarding itself.
//
// At most one tour comes back. Two overlays on screen at the same time is
// broken UX, so ties are resolved by priority and then age — deterministic, and
// configurable from the admin panel.
func (rd *RuntimeDomain) Resolve(ctx context.Context, req ResolveRequest) (*ResolveResult, error) {
	if req.SubjectId == "" {
		return nil, logicErr(ErrSubjectRequired, "resolve", http.StatusBadRequest)
	}

	candidates, err := rd.tours.ResolveCandidates(ctx, req.AppId)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	progress, err := rd.runtime.SubjectProgress(ctx, req.AppId, req.SubjectId)
	if err != nil {
		return nil, err
	}

	path := PathFromURL(req.Url)

	// Candidates arrive already ordered by priority DESC, created_at ASC, so the
	// first survivor of the filters is the winner.
	for _, tour := range candidates {
		if !MatchPath(tour.TargetPath, path) {
			continue
		}
		if !matchAudience(tour.Audience, req.Props, progress[tour.Id]) {
			continue
		}
		return rd.start(ctx, req, tour, progress[tour.Id])
	}

	return nil, nil
}

// start reads or creates the progress row and returns the tour to show.
func (rd *RuntimeDomain) start(ctx context.Context, req ResolveRequest, tour *Tour, current *Progress) (*ResolveResult, error) {
	hints, err := rd.hints.ListByVersion(ctx, tour.VersionId)
	if err != nil {
		return nil, err
	}
	tour.Hints = hints

	now := time.Now().UTC()

	switch {
	case current == nil:
		current = &Progress{
			AppId:         req.AppId,
			SubjectId:     req.SubjectId,
			TourId:        tour.Id,
			TourVersionId: tour.VersionId,
			Status:        ProgressInProgress,
			ShowsCount:    1,
			StartedAt:     now,
		}

	case current.TourVersionId != tour.VersionId:
		// The subject was mid-way through a version that is no longer published.
		// The position is not carried over: hint ids in the new version are
		// different rows entirely, and guessing an equivalent would produce
		// events that belong to no version anyone ever saw. Starting over is the
		// honest option.
		current = &Progress{
			AppId:         req.AppId,
			SubjectId:     req.SubjectId,
			TourId:        tour.Id,
			TourVersionId: tour.VersionId,
			Status:        ProgressInProgress,
			ShowsCount:    current.ShowsCount + 1,
			StartedAt:     now,
		}

	default:
		current.ShowsCount++
	}

	if current.Status == ProgressInProgress && current.CurrentHintId == nil && len(hints) > 0 {
		current.CurrentHintId = &hints[0].Id
	}

	if err := rd.runtime.UpsertProgress(ctx, current); err != nil {
		return nil, err
	}

	return &ResolveResult{
		Tour:          tour,
		TourVersionId: tour.VersionId,
		Version:       tour.Version,
		CurrentHintId: current.CurrentHintId,
	}, nil
}

// matchAudience evaluates targeting on the backend. It is not sent to the
// browser to be evaluated there, because targeting rules in the browser are
// targeting rules the browser can rewrite.
func matchAudience(a Audience, props map[string]any, p *Progress) bool {
	if a.OnlyNew && !boolProp(props, "isNewUser") {
		return false
	}
	if p == nil {
		return true
	}
	if a.ShowOnce && (p.Status == ProgressCompleted || p.Status == ProgressDismissed) {
		return false
	}
	if a.MaxShows > 0 && p.ShowsCount >= a.MaxShows {
		return false
	}
	return true
}

func boolProp(props map[string]any, key string) bool {
	v, ok := props[key]
	if !ok {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

// Ingest accepts a batch of events.
//
// Each event is checked individually and a bad one is counted in Rejected
// instead of failing the batch: the SDK sends up to fifty at a time, and one
// malformed entry from an older client version must not discard the other
// forty-nine. This is also why the event vocabulary is not a CHECK constraint —
// a constraint violation would take the whole transaction with it.
func (rd *RuntimeDomain) Ingest(
	ctx context.Context,
	appId uuid.UUID,
	subjectId, sessionId string,
	events []Event,
) (*EventBatchResult, error) {
	switch {
	case subjectId == "":
		return nil, logicErr(ErrSubjectRequired, "events", http.StatusBadRequest)
	case sessionId == "":
		return nil, logicErr(ErrSessionRequired, "events", http.StatusBadRequest)
	case len(events) == 0:
		return nil, logicErr(ErrEmptyBatch, "events", http.StatusBadRequest)
	case len(events) > maxEventsPerBatch:
		return nil, logicErr(ErrBatchTooLarge, "events", http.StatusBadRequest)
	}

	result := &EventBatchResult{}
	valid := make([]Event, 0, len(events))

	for i := range events {
		e := events[i]
		e.AppId = appId
		e.SubjectId = subjectId
		e.SessionId = sessionId
		if e.OccurredAt.IsZero() {
			e.OccurredAt = time.Now().UTC()
		}
		if e.Payload == nil {
			e.Payload = map[string]any{}
		}

		if err := validateEvent(&e); err != nil {
			result.Rejected++
			result.Errors = append(result.Errors, err.Error())
			continue
		}
		valid = append(valid, e)
	}

	if len(valid) > 0 {
		accepted, duplicates, err := rd.runtime.InsertEvents(ctx, valid)
		if err != nil {
			return nil, err
		}
		result.Accepted = accepted
		result.Duplicates = duplicates
	}

	return result, nil
}

func validateEvent(e *Event) error {
	if e.EventKey == "" {
		return ErrEmptyBatch
	}
	if !e.Type.Valid() {
		return ErrUnknownEventType
	}
	if e.TourVersionId == uuid.Nil {
		// Without the version an event is unusable: impressions from before and
		// after an edit would land in the same bucket with nothing to tell them
		// apart. This is the one rule the analytics rest on.
		return ErrHintIdMismatch
	}
	if e.Type.TourLevel() != (e.HintId == nil) {
		return ErrHintIdMismatch
	}
	return nil
}

// Analytics builds the funnel for one version over one period.
func (rd *RuntimeDomain) Analytics(
	ctx context.Context,
	tourId, versionId uuid.UUID,
	from, to time.Time,
) (*Analytics, error) {
	if !to.After(from) {
		return nil, logicErr(ErrBadPeriod, "analytics", http.StatusBadRequest)
	}

	if versionId == uuid.Nil {
		published, err := rd.tours.VersionByStatus(ctx, tourId, TourPublished)
		if err != nil {
			return nil, err
		}
		versionId = published.Id
	}

	version, err := rd.tours.GetVersion(ctx, versionId)
	if err != nil {
		return nil, err
	}
	if version.TourId != tourId {
		return nil, logicErr(ErrVersionNotInTour, "analytics", http.StatusBadRequest)
	}

	funnel, err := rd.runtime.Funnel(ctx, versionId, from, to)
	if err != nil {
		return nil, err
	}
	totals, err := rd.runtime.Totals(ctx, versionId, from, to)
	if err != nil {
		return nil, err
	}

	if totals.Started > 0 {
		totals.CompletionRate = round3(float64(totals.Completed) / float64(totals.Started))
		totals.GoalRate = round3(float64(totals.GoalReached) / float64(totals.Started))
	}

	broken := make([]uuid.UUID, 0)
	for i := range funnel {
		if funnel[i].Shown > 0 {
			funnel[i].Dropoff = round3(float64(funnel[i].Shown-funnel[i].Completed) / float64(funnel[i].Shown))
		}
		// A missing anchor is a product signal, not a debug line: it means the
		// host's markup moved and the tour has quietly stopped working.
		if funnel[i].SelectorMissing > 0 {
			broken = append(broken, funnel[i].HintId)
		}
	}

	return &Analytics{
		TourId:          tourId,
		TourVersionId:   versionId,
		Version:         version.Version,
		From:            from,
		To:              to,
		Totals:          totals,
		Funnel:          funnel,
		BrokenSelectors: broken,
	}, nil
}

func round3(v float64) float64 {
	return float64(int64(v*1000+0.5)) / 1000
}
