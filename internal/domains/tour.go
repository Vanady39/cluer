package domains

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type (
	TourDomain struct {
		tours TourRepositoryInterface
		hints HintRepositoryInterface
	}

	TourDomainInterface interface {
		Create(ctx context.Context, appId uuid.UUID, t *Tour) (*Tour, error)
		List(ctx context.Context, appId uuid.UUID) ([]*Tour, error)
		Card(ctx context.Context, tourId uuid.UUID) (*TourCard, error)
		UpdateMeta(ctx context.Context, tourId uuid.UUID, p TourMetaPatch) (*Tour, error)
		Archive(ctx context.Context, tourId uuid.UUID) error

		ListVersions(ctx context.Context, tourId uuid.UUID) ([]*TourVersion, error)
		GetVersion(ctx context.Context, versionId uuid.UUID) (*TourVersion, error)
		CreateDraft(ctx context.Context, tourId uuid.UUID) (*TourVersion, error)
		UpdateDraft(ctx context.Context, tourId uuid.UUID, p DraftPatch) (*TourVersion, error)
		Publish(ctx context.Context, tourId uuid.UUID) (*TourVersion, error)
		Rollback(ctx context.Context, tourId, toVersionId uuid.UUID) (*TourVersion, error)

		ListPublished(ctx context.Context, appId uuid.UUID, path string) ([]*Tour, error)
	}

	TourRepositoryInterface interface {
		CreateWithDraft(ctx context.Context, t *Tour) (*Tour, error)
		GetById(ctx context.Context, id uuid.UUID) (*Tour, error)
		List(ctx context.Context, appId uuid.UUID) ([]*Tour, error)
		UpdateMeta(ctx context.Context, id uuid.UUID, p TourMetaPatch) (*Tour, error)
		Archive(ctx context.Context, id uuid.UUID) error

		GetVersion(ctx context.Context, versionId uuid.UUID) (*TourVersion, error)
		VersionOwner(ctx context.Context, versionId uuid.UUID) (tourId, appId uuid.UUID, err error)
		ListVersions(ctx context.Context, tourId uuid.UUID) ([]*TourVersion, error)
		VersionByStatus(ctx context.Context, tourId uuid.UUID, s TourStatus) (*TourVersion, error)
		UpdateDraft(ctx context.Context, versionId uuid.UUID, p DraftPatch) (*TourVersion, error)
		CopyToDraft(ctx context.Context, tourId, sourceVersionId uuid.UUID) (*TourVersion, error)
		Publish(ctx context.Context, tourId uuid.UUID) (*TourVersion, error)

		ResolveCandidates(ctx context.Context, appId uuid.UUID) ([]*Tour, error)
	}

	TourMetaPatch struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Enabled     *bool   `json:"enabled"`
		Priority    *int    `json:"priority"`
	}

	DraftPatch struct {
		TriggerType   *TriggerType   `json:"trigger_type"`
		TriggerConfig *TriggerConfig `json:"trigger_config"`
		TargetPath    *string        `json:"target_path"`
		Audience      *Audience      `json:"audience"`
	}
)

func NewTourDomain(tours TourRepositoryInterface, hints HintRepositoryInterface) *TourDomain {
	return &TourDomain{tours: tours, hints: hints}
}

func (td *TourDomain) Create(ctx context.Context, appId uuid.UUID, t *Tour) (*Tour, error) {
	if err := validateTour(t); err != nil {
		return nil, err
	}

	t.AppId = appId
	t.Status = TourDraft
	if t.TriggerType == "" {
		t.TriggerType = TriggerOnLoad
	}
	if t.Audience == nil {
		t.Audience = &Audience{}
	}

	if details := validateDraftBody(t.TriggerType, t.TriggerConfig, t.Audience.Rules); len(details) > 0 {
		return nil, &ValidationError{Err: ErrDraftInvalid, Details: details}
	}

	return td.tours.CreateWithDraft(ctx, t)
}

func (td *TourDomain) List(ctx context.Context, appId uuid.UUID) ([]*Tour, error) {
	return td.tours.List(ctx, appId)
}

func (td *TourDomain) Card(ctx context.Context, tourId uuid.UUID) (*TourCard, error) {
	tour, err := td.tours.GetById(ctx, tourId)
	if err != nil {
		return nil, err
	}

	published, err := td.versionWithHints(ctx, tourId, TourPublished)
	if err != nil {
		return nil, err
	}
	draft, err := td.versionWithHints(ctx, tourId, TourDraft)
	if err != nil {
		return nil, err
	}

	return &TourCard{Tour: tour, Published: published, Draft: draft}, nil
}

func (td *TourDomain) versionWithHints(ctx context.Context, tourId uuid.UUID, s TourStatus) (*TourVersion, error) {
	v, err := td.versionOrNil(ctx, tourId, s)
	if err != nil || v == nil {
		return nil, err
	}
	hints, err := td.hints.ListByVersion(ctx, v.Id)
	if err != nil {
		return nil, err
	}
	v.Hints = hints
	return v, nil
}

func (td *TourDomain) UpdateMeta(ctx context.Context, tourId uuid.UUID, p TourMetaPatch) (*Tour, error) {
	if p.Title != nil && *p.Title == "" {
		return nil, logicErr(ErrTitleRequired, "tour update", http.StatusBadRequest)
	}
	if p.Priority != nil && *p.Priority < 0 {
		return nil, logicErr(ErrPriorityNegative, "tour update", http.StatusBadRequest)
	}
	return td.tours.UpdateMeta(ctx, tourId, p)
}

func (td *TourDomain) Archive(ctx context.Context, tourId uuid.UUID) error {
	return td.tours.Archive(ctx, tourId)
}

func (td *TourDomain) ListVersions(ctx context.Context, tourId uuid.UUID) ([]*TourVersion, error) {
	return td.tours.ListVersions(ctx, tourId)
}

func (td *TourDomain) GetVersion(ctx context.Context, versionId uuid.UUID) (*TourVersion, error) {
	version, err := td.tours.GetVersion(ctx, versionId)
	if err != nil {
		return nil, err
	}

	hints, err := td.hints.ListByVersion(ctx, versionId)
	if err != nil {
		return nil, err
	}
	version.Hints = hints

	return version, nil
}

func (td *TourDomain) CreateDraft(ctx context.Context, tourId uuid.UUID) (*TourVersion, error) {
	if existing, err := td.versionOrNil(ctx, tourId, TourDraft); err != nil {
		return nil, err
	} else if existing != nil {
		return nil, logicErr(ErrDraftAlreadyExists, "draft creation", http.StatusConflict)
	}

	source, err := td.versionOrNil(ctx, tourId, TourPublished)
	if err != nil {
		return nil, err
	}
	if source == nil {
		return nil, logicErr(ErrNoVersionToCopy, "draft creation", http.StatusConflict)
	}

	return td.tours.CopyToDraft(ctx, tourId, source.Id)
}

func (td *TourDomain) UpdateDraft(ctx context.Context, tourId uuid.UUID, p DraftPatch) (*TourVersion, error) {
	draft, err := td.requireDraft(ctx, tourId)
	if err != nil {
		return nil, err
	}
	if p.TriggerType != nil && !p.TriggerType.Valid() {
		return nil, logicErr(ErrInvalidTriggerType, "draft update", http.StatusBadRequest)
	}

	triggerType := draft.TriggerType
	if p.TriggerType != nil {
		triggerType = *p.TriggerType
	}

	config := draft.TriggerConfig
	if p.TriggerConfig != nil {
		config = p.TriggerConfig
	}

	rules := draft.Audience.Rules
	if p.Audience != nil {
		rules = p.Audience.Rules
	}

	if details := validateDraftBody(triggerType, config, rules); len(details) > 0 {
		return nil, &ValidationError{Err: ErrDraftInvalid, Details: details}
	}
	return td.tours.UpdateDraft(ctx, draft.Id, p)
}

func (td *TourDomain) Publish(ctx context.Context, tourId uuid.UUID) (*TourVersion, error) {
	draft, err := td.requireDraft(ctx, tourId)
	if err != nil {
		return nil, err
	}

	hints, err := td.hints.ListByVersion(ctx, draft.Id)
	if err != nil {
		return nil, err
	}
	if err := validateForPublish(draft, hints); err != nil {
		return nil, err
	}

	return td.tours.Publish(ctx, tourId)
}

func (td *TourDomain) Rollback(ctx context.Context, tourId, toVersionId uuid.UUID) (*TourVersion, error) {
	source, err := td.tours.GetVersion(ctx, toVersionId)
	if err != nil {
		return nil, err
	}
	if source.TourId != tourId {
		return nil, logicErr(ErrVersionNotInTour, "rollback", http.StatusBadRequest)
	}

	if draft, err := td.versionOrNil(ctx, tourId, TourDraft); err != nil {
		return nil, err
	} else if draft != nil {
		return nil, logicErr(ErrDraftBlocksRollback, "rollback", http.StatusConflict)
	}

	if _, err := td.tours.CopyToDraft(ctx, tourId, toVersionId); err != nil {
		return nil, err
	}
	return td.tours.Publish(ctx, tourId)
}

func (td *TourDomain) ListPublished(ctx context.Context, appId uuid.UUID, path string) ([]*Tour, error) {
	candidates, err := td.tours.ResolveCandidates(ctx, appId)
	if err != nil {
		return nil, err
	}

	matched := make([]*Tour, 0, len(candidates))
	for _, c := range candidates {
		if MatchPath(c.TargetPath, path) {
			hints, err := td.hints.ListByVersion(ctx, c.VersionId)
			if err != nil {
				return nil, err
			}
			c.Hints = hints
			matched = append(matched, c)
		}
	}
	return matched, nil
}

func (td *TourDomain) versionOrNil(ctx context.Context, tourId uuid.UUID, s TourStatus) (*TourVersion, error) {
	v, err := td.tours.VersionByStatus(ctx, tourId, s)
	if err != nil {
		if IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return v, nil
}

func (td *TourDomain) requireDraft(ctx context.Context, tourId uuid.UUID) (*TourVersion, error) {
	draft, err := td.versionOrNil(ctx, tourId, TourDraft)
	if err != nil {
		return nil, err
	}
	if draft == nil {
		return nil, logicErr(ErrNoDraft, "draft lookup", http.StatusConflict)
	}
	return draft, nil
}

func validateTour(t *Tour) error {
	var err error
	switch {
	case t.Title == "":
		err = ErrTitleRequired
	case t.TargetPath == "":
		err = ErrTargetPathRequired
	case t.Priority < 0:
		err = ErrPriorityNegative
	case t.TriggerType != "" && !t.TriggerType.Valid():
		err = ErrInvalidTriggerType
	case t.Audience != nil && t.Audience.MaxShows < 0:
		err = ErrMaxShowsNegative
	default:
		return nil
	}
	return logicErr(err, "tour validation", http.StatusBadRequest)
}

// Проверяется строго то поле, которое нужно текущему trigger_type. Конфиг
// присылается целиком и при смене типа тащит за собой поля прежней формы —
// ругаться на них значит отдавать фронту 400 по полю, которого в форме уже нет.
// Путь ошибки всегда указывает на конкретное поле, даже когда конфига нет
// вовсе: для формы это одна и та же ситуация «поле не заполнено».
func validateTriggerConfig(tType TriggerType, cfg *TriggerConfig) []ErrorDetail {
	detail := func(field, message string) []ErrorDetail {
		return []ErrorDetail{{Path: "trigger_config." + field, Message: message}}
	}

	switch tType {
	case TriggerDelay:
		if cfg == nil || cfg.DelayMs == nil {
			return detail("delay_ms", "is required when trigger_type is delay")
		}
		if *cfg.DelayMs < 100 || *cfg.DelayMs > 600000 {
			return detail("delay_ms", "must be an integer between 100 and 600000 (0.1s to 10min)")
		}

	case TriggerScrollDepth:
		if cfg == nil || cfg.ScrollDepth == nil {
			return detail("scroll_depth", "is required when trigger_type is scroll_depth")
		}
		if *cfg.ScrollDepth < 1 || *cfg.ScrollDepth > 100 {
			return detail("scroll_depth", "must be an integer between 1 and 100")
		}

	case TriggerInactivity:
		if cfg == nil || cfg.InactivitySecs == nil {
			return detail("inactivity_secs", "is required when trigger_type is inactivity")
		}
		if *cfg.InactivitySecs < 3 {
			return detail("inactivity_secs", "must be at least 3 seconds to avoid accidental triggers")
		}

	case TriggerElementVisible:
		if cfg == nil || cfg.ElementSelector == nil || strings.TrimSpace(*cfg.ElementSelector) == "" {
			return detail("element_selector",
				"is required and must not be empty when trigger_type is element_visible")
		}
	}

	return nil
}

// Тело версии проверяется одним набором правил на всех трёх входах: создание
// тура, правка черновика и публикация. Иначе админка получает 201 на то, что
// потом не публикуется, и узнаёт о проблеме на шаг позже, чем могла бы.
func validateDraftBody(tType TriggerType, cfg *TriggerConfig, rules []AudienceRule) []ErrorDetail {
	details := validateTriggerConfig(tType, cfg)
	return append(details, validateAudienceRules(rules)...)
}

func validateForPublish(v *TourVersion, hints []Hint) error {
	details := make([]ErrorDetail, 0)

	if v.TargetPath == "" {
		details = append(details, ErrorDetail{Path: "target_path", Message: "must not be empty"})
	}
	if !v.TriggerType.Valid() {
		details = append(details, ErrorDetail{Path: "trigger_type", Message: "unknown trigger type"})
	}
	details = append(details, validateDraftBody(v.TriggerType, v.TriggerConfig, v.Audience.Rules)...)
	if len(hints) == 0 {
		details = append(details, ErrorDetail{Path: "hints", Message: "a tour with no hints has nothing to show"})
	}

	seen := make(map[int]bool, len(hints))
	for i, h := range hints {
		field := func(name string) string {
			return "hints[" + strconv.Itoa(i) + "]." + name
		}
		if h.Title == "" {
			details = append(details, ErrorDetail{Path: field("title"), Message: "must not be empty"})
		}
		if h.Content == "" {
			details = append(details, ErrorDetail{Path: field("content"), Message: "must not be empty"})
		}
		if !h.Placement.Valid() {
			details = append(details, ErrorDetail{Path: field("placement"), Message: "unknown placement"})
		}
		if h.Placement != PlacementCenter && h.Selector == "" {
			details = append(details, ErrorDetail{Path: field("selector"), Message: "required unless placement is center"})
		}
		// Пустой путь — законное «наследую target_path». А вот относительный
		// никогда не совпадёт с location.pathname, и тур молча встанет на этом
		// шаге, дожидаясь элемента со страницы, куда он не перешёл.
		if h.PagePath != "" && !strings.HasPrefix(h.PagePath, "/") {
			details = append(details, ErrorDetail{Path: field("page_path"), Message: "must start with /"})
		}
		if seen[h.Step] {
			details = append(details, ErrorDetail{Path: field("step"), Message: "duplicate step number"})
		}
		seen[h.Step] = true
	}

	for n := 1; n <= len(hints); n++ {
		if !seen[n] {
			details = append(details, ErrorDetail{Path: "hints", Message: "step numbers must be contiguous starting at 1"})
			break
		}
	}

	if len(details) > 0 {
		return &ValidationError{Err: ErrTourNotPublishable, Details: details}
	}
	return nil
}

func validateAudienceRules(rules []AudienceRule) []ErrorDetail {
	var details []ErrorDetail

	validTypes := map[string]bool{
		"event_performed": true,
		"page_visited":    true,
		"user_property":   true,
	}
	validOperators := map[string]bool{
		"exists": true, "not_exists": true,
		"eq": true, "neq": true,
		"gt": true, "gte": true, "lt": true, "lte": true,
		"contains": true, "not_contains": true,
		"starts_with": true, "ends_with": true,
	}
	existenceOnly := map[string]bool{
		"event_performed": true,
		"page_visited":    true,
	}
	for i, r := range rules {
		prefix := "audience.rules[" + strconv.Itoa(i) + "]."

		if r.Type == "" {
			details = append(details, ErrorDetail{
				Path:    prefix + "type",
				Message: "is required",
			})
		} else if !validTypes[r.Type] {
			details = append(details, ErrorDetail{
				Path:    prefix + "type",
				Message: fmt.Sprintf("unknown type %q (expected event_performed, page_visited, or user_property)", r.Type),
			})
		}
		if r.Key == "" {
			details = append(details, ErrorDetail{
				Path:    prefix + "key",
				Message: "is required",
			})
		}

		// Оператор проверяем один раз и запоминаем результат: жаловаться ещё и на
		// value, когда оператора нет или он неизвестен, — лишний шум для формы.
		operatorOk := false
		if r.Operator == "" {
			details = append(details, ErrorDetail{
				Path:    prefix + "operator",
				Message: "is required",
			})
		} else if !validOperators[r.Operator] {
			details = append(details, ErrorDetail{
				Path:    prefix + "operator",
				Message: fmt.Sprintf("unknown operator %q", r.Operator),
			})
		} else if existenceOnly[r.Type] && r.Operator != "exists" && r.Operator != "not_exists" {
			details = append(details, ErrorDetail{
				Path:    prefix + "operator",
				Message: fmt.Sprintf("must be exists or not_exists for %s rules", r.Type),
			})
		} else {
			operatorOk = true
		}

		if r.Type == "event_performed" || r.Type == "page_visited" {
			if _, ok := parseTimeframe(r.Timeframe); !ok {
				details = append(details, ErrorDetail{
					Path:    prefix + "timeframe",
					Message: fmt.Sprintf("unknown timeframe %q (expected 1h, 24h, 7d, 30d, or all_time)", r.Timeframe),
				})
			}
		}
		if r.Type == "user_property" && r.Timeframe != "" {
			details = append(details, ErrorDetail{
				Path:    prefix + "timeframe",
				Message: "must be empty for user_property rules",
			})
		}

		if r.Type == "user_property" && operatorOk &&
			r.Operator != "exists" && r.Operator != "not_exists" && r.Value == nil {
			details = append(details, ErrorDetail{
				Path:    prefix + "value",
				Message: "is required for comparison operators",
			})
		}
	}

	return details
}
