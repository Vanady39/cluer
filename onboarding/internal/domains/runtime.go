package domains

import (
	"context"
	"fmt"
	"github.com/Vanady39/cluer/platform/errs"
	"net/http"
	"strconv"
	"strings"
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
		GetSubjectEvents(ctx context.Context, appId uuid.UUID, subjectId string, minSince time.Time, limit int) ([]Event, error)

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
	unbounded := false
	var minSince time.Time
	needHistory := false
	for _, tour := range candidates {
		if tour.Audience == nil || len(tour.Audience.Rules) == 0 {
			continue
		}

		needHistory = true

		since := computeMinSince(tour.Audience.Rules)
		if since.IsZero() {
			unbounded = true
			break
		}
		if minSince.IsZero() || since.Before(minSince) {
			minSince = since
		}
	}

	var eventHistory []Event
	if needHistory {
		if unbounded {
			minSince = time.Time{}
		}
		eventHistory, err = rd.runtime.GetSubjectEvents(ctx, req.AppId, req.SubjectId, minSince, 1000)
		if err != nil {
			return nil, err
		}
	}

	for _, tour := range candidates {
		if !MatchPath(tour.TargetPath, path) {
			continue
		}

		if tour.Audience != nil {
			if !matchAudience(*tour.Audience, req.Props, progress[tour.Id]) {
				continue
			}
			if len(tour.Audience.Rules) > 0 {
				if !matchAudienceRules(tour.Audience.Rules, eventHistory, req.Props) {
					continue
				}
			}
		}
		return rd.start(ctx, req, tour, progress[tour.Id])
	}
	return nil, nil
}

func computeMinSince(rules []AudienceRule) time.Time {
	now := time.Now()
	var min time.Time
	for _, r := range rules {
		if r.Type == "user_property" {
			continue
		}
		duration, ok := parseTimeframe(r.Timeframe)
		if !ok {
			return time.Time{}
		}
		if duration == 0 {
			return time.Time{}
		}
		t := now.Add(-duration)
		if min.IsZero() || t.Before(min) {
			min = t
		}
	}
	return min
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

func applyExistenceOperator(op string, found bool) bool {
	switch op {
	case "not_exists":
		return !found
	case "exists", "":
		return found
	default:
		return false
	}
}

func matchAudienceRules(rules []AudienceRule, history []Event, props map[string]any) bool {
	for _, rule := range rules {
		matched := false

		switch rule.Type {
		case "event_performed":
			found := false
			for _, evt := range history {
				// Матчим строго по payload.event_name: фолбэк на evt.Type открыл бы
				// таргетинг по служебным типам SDK (tour_completed, hint_shown), чего
				// контракт не обещает.
				evtName, ok := evt.Payload["event_name"].(string)
				if !ok {
					continue
				}
				if evtName == rule.Key && checkTimeframe(evt.OccurredAt, rule.Timeframe) {
					found = true
					break
				}
			}
			matched = applyExistenceOperator(rule.Operator, found)

		case "page_visited":
			found := false
			for _, evt := range history {
				isPageView := string(evt.Type) == "page_view" || evt.Payload["event_name"] == "page_view"
				if isPageView {
					if url, ok := evt.Payload["url"].(string); ok && url == rule.Key {
						if checkTimeframe(evt.OccurredAt, rule.Timeframe) {
							found = true
							break
						}
					}
				}
			}
			matched = applyExistenceOperator(rule.Operator, found)

		case "user_property":
			val, exists := props[rule.Key]
			switch rule.Operator {
			case "exists":
				matched = exists
			case "not_exists":
				matched = !exists
			default:
				if exists {
					matched = evaluateOperator(rule.Operator, val, rule.Value)
				}
			}
		default:
			matched = false
		}
		if !matched {
			return false
		}
	}
	return true
}

func parseTimeframe(tf string) (time.Duration, bool) {
	switch tf {
	case "", "all_time":
		return 0, true
	case "1h":
		return time.Hour, true
	case "24h":
		return 24 * time.Hour, true
	case "7d":
		return 7 * 24 * time.Hour, true
	case "30d":
		return 30 * 24 * time.Hour, true
	default:
		return 0, false
	}
}

func checkTimeframe(eventTime time.Time, timeframe string) bool {
	duration, ok := parseTimeframe(timeframe)
	if !ok {
		return false
	}
	if duration == 0 {
		return true
	}
	return time.Since(eventTime) <= duration
}

func evaluateOperator(op string, actual any, expected any) bool {
	if actual == nil && expected == nil {
		return op == "eq"
	}
	if actual == nil || expected == nil {
		return op == "neq"
	}
	switch op {
	case "eq":
		return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected)
	case "neq":
		return fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", expected)
	}

	actualFloat, actualIsNum := toFloat64(actual)
	expectedFloat, expectedIsNum := toFloat64(expected)

	if actualIsNum && expectedIsNum {
		switch op {
		case "gt":
			return actualFloat > expectedFloat
		case "gte":
			return actualFloat >= expectedFloat
		case "lt":
			return actualFloat < expectedFloat
		case "lte":
			return actualFloat <= expectedFloat
		}
	}

	actualStr, actualIsStr := actual.(string)
	expectedStr, expectedIsStr := expected.(string)

	if actualIsStr && expectedIsStr {
		switch op {
		case "contains":
			return strings.Contains(actualStr, expectedStr)
		case "not_contains":
			return !strings.Contains(actualStr, expectedStr)
		case "starts_with":
			return strings.HasPrefix(actualStr, expectedStr)
		case "ends_with":
			return strings.HasSuffix(actualStr, expectedStr)
		case "gt":
			return actualStr > expectedStr
		case "gte":
			return actualStr >= expectedStr
		case "lt":
			return actualStr < expectedStr
		case "lte":
			return actualStr <= expectedStr
		}
	}
	if actualSlice, ok := actual.([]any); ok {
		expectedStrFmt := fmt.Sprintf("%v", expected)
		switch op {
		case "contains":
			for _, item := range actualSlice {
				if fmt.Sprintf("%v", item) == expectedStrFmt {
					return true
				}
			}
			return false
		case "not_contains":
			for _, item := range actualSlice {
				if fmt.Sprintf("%v", item) == expectedStrFmt {
					return false
				}
			}
			return true
		}
	}

	return false
}

func toFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, true
		}
	}
	return 0, false
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

		var scope *eventScope
		if e.TourVersionId != uuid.Nil {
			s, err := rd.scopeOf(ctx, scopes, e.TourVersionId)
			if err != nil {
				if errs.IsNotFound(err) {
					result.reject(ErrVersionNotFound)
					continue
				}
				return nil, err
			}
			scope = s
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
			return nil, errs.NotFound(ErrVersionNotFound, versionId.String())
		}
		return cached, nil
	}

	tourId, appId, err := rd.tours.VersionOwner(ctx, versionId)
	if err != nil {
		if errs.IsNotFound(err) {
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
	if e.Type == EventCustom && e.TourVersionId == uuid.Nil {
		if e.TourId != uuid.Nil {
			return ErrTourUnexpected
		}
		if e.HintId != nil {
			return ErrHintUnexpected
		}
		return nil
	}
	if e.TourVersionId == uuid.Nil {
		return ErrVersionRequired
	}
	if e.Type != EventCustom && e.Type.TourLevel() != (e.HintId == nil) {
		return ErrHintIdMismatch
	}
	return nil
}

func validateEventScope(e *Event, appId uuid.UUID, scope *eventScope) error {
	if e.Type == EventCustom && e.TourVersionId == uuid.Nil {
		return nil
	}
	if scope == nil || scope.AppId != appId {
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
