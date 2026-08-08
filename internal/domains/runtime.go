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

	for _, tour := range candidates {
		if !MatchPath(tour.TargetPath, path) {
			continue
		}
		if tour.Audience != nil && !matchAudience(*tour.Audience, req.Props, progress[tour.Id]) {
			continue
		}
		return rd.start(ctx, req, tour, progress[tour.Id])
	}

	return nil, nil
}

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
	scopes := make(map[uuid.UUID]*eventScope)

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
			result.reject(err)
			continue
		}

		scope, err := rd.scopeOf(ctx, scopes, e.TourVersionId)
		if err != nil {
			if IsNotFound(err) {
				result.reject(ErrVersionNotFound)
				continue
			}
			return nil, err
		}
		if err := validateEventScope(&e, appId, scope); err != nil {
			result.reject(err)
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

type eventScope struct {
	TourId uuid.UUID
	AppId  uuid.UUID
	Hints  map[uuid.UUID]struct{}
}

func (rd *RuntimeDomain) scopeOf(
	ctx context.Context,
	cache map[uuid.UUID]*eventScope,
	versionId uuid.UUID,
) (*eventScope, error) {
	if cached, seen := cache[versionId]; seen {
		if cached == nil {
			return nil, NotFound(ErrVersionNotFound, versionId.String())
		}
		return cached, nil
	}

	tourId, appId, err := rd.tours.VersionOwner(ctx, versionId)
	if err != nil {
		if IsNotFound(err) {
			cache[versionId] = nil
		}
		return nil, err
	}

	hints, err := rd.hints.ListByVersion(ctx, versionId)
	if err != nil {
		return nil, err
	}

	ids := make(map[uuid.UUID]struct{}, len(hints))
	for _, h := range hints {
		ids[h.Id] = struct{}{}
	}

	scope := &eventScope{TourId: tourId, AppId: appId, Hints: ids}
	cache[versionId] = scope
	return scope, nil
}

func validateEvent(e *Event) error {
	if e.EventKey == "" {
		return ErrEventKeyRequired
	}
	if !e.Type.Valid() {
		return ErrUnknownEventType
	}
	if e.TourVersionId == uuid.Nil {
		return ErrVersionRequired
	}
	if e.Type.TourLevel() != (e.HintId == nil) {
		return ErrHintIdMismatch
	}
	return nil
}

func validateEventScope(e *Event, appId uuid.UUID, scope *eventScope) error {
	if scope.AppId != appId {
		return ErrVersionForeignApp
	}
	if e.TourId == uuid.Nil {
		return ErrTourRequired
	}
	if e.TourId != scope.TourId {
		return ErrTourVersionMismatch
	}
	if e.HintId != nil {
		if _, ok := scope.Hints[*e.HintId]; !ok {
			return ErrHintNotInVersion
		}
	}
	return nil
}

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
