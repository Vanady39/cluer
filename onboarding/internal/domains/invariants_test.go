package domains

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchPath(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		{"звёздочка ловит всё", "/*", "/anything/deep", true},
		{"точное совпадение", "/dashboard", "/dashboard", true},
		{"точное не ловит вложенное", "/dashboard", "/dashboard/stats", false},
		{"хвостовой слеш игнорируется", "/dashboard", "/dashboard/", true},
		{"префикс со звёздочкой", "/additem*", "/additem/new", true},
		{"префикс не ловит чужое", "/additem*", "/profile", false},
		{"суффикс со звёздочкой", "*/edit", "/items/42/edit", true},
		{"звёздочка в середине", "/items/*/edit", "/items/42/edit", true},
		{"query игнорируется, если не упомянут", "/dashboard", "/dashboard?tab=stats", true},
		{"query учитывается, если упомянут", "/dashboard?tab=stats", "/dashboard?tab=stats", true},
		{"пустой паттерн не ловит ничего", "", "/dashboard", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, MatchPath(tt.pattern, tt.path))
		})
	}
}

func TestPathFromURL(t *testing.T) {
	assert.Equal(t, "/additem?cat=1", PathFromURL("https://demo.local/additem?cat=1"))
	assert.Equal(t, "/dashboard", PathFromURL("http://localhost:3000/dashboard"))
	assert.Equal(t, "/", PathFromURL("https://demo.local"))
}

func TestValidateForPublish(t *testing.T) {
	validHint := func(step int) Hint {
		return Hint{
			Step: step, Title: "t", Content: "c",
			Selector: "#id", Placement: PlacementBottom,
		}
	}
	version := func() *TourVersion {
		return &TourVersion{TargetPath: "/dashboard", TriggerType: TriggerOnLoad}
	}

	t.Run("валидный тур публикуется", func(t *testing.T) {
		assert.NoError(t, validateForPublish(version(), []Hint{validHint(1), validHint(2)}))
	})

	t.Run("тур без подсказок не публикуется", func(t *testing.T) {
		err := validateForPublish(version(), nil)
		require.Error(t, err)
		assert.Contains(t, detailPaths(t, err), "hints")
	})

	t.Run("дыра в нумерации ловится", func(t *testing.T) {
		err := validateForPublish(version(), []Hint{validHint(1), validHint(3)})
		require.Error(t, err)
		assert.Contains(t, detailPaths(t, err), "hints")
	})

	t.Run("дубликат шага ловится", func(t *testing.T) {
		err := validateForPublish(version(), []Hint{validHint(1), validHint(1)})
		require.Error(t, err)
		assert.Contains(t, detailPaths(t, err), "hints[1].step")
	})

	t.Run("пустой page_path допустим — наследует target_path", func(t *testing.T) {
		assert.NoError(t, validateForPublish(version(), []Hint{validHint(1)}))
	})

	t.Run("абсолютный page_path допустим", func(t *testing.T) {
		crossPage := validHint(1)
		crossPage.PagePath = "/profile"
		assert.NoError(t, validateForPublish(version(), []Hint{crossPage}))
	})

	t.Run("относительный page_path ловится", func(t *testing.T) {
		relative := validHint(1)
		relative.PagePath = "profile"
		err := validateForPublish(version(), []Hint{relative})
		require.Error(t, err)
		assert.Contains(t, detailPaths(t, err), "hints[0].page_path")
	})

	t.Run("селектор обязателен, кроме placement=center", func(t *testing.T) {
		noSelector := validHint(1)
		noSelector.Selector = ""
		err := validateForPublish(version(), []Hint{noSelector})
		require.Error(t, err)
		assert.Contains(t, detailPaths(t, err), "hints[0].selector")

		centered := noSelector
		centered.Placement = PlacementCenter
		assert.NoError(t, validateForPublish(version(), []Hint{centered}))
	})

	t.Run("все проблемы возвращаются разом", func(t *testing.T) {
		broken := Hint{Step: 1}
		err := validateForPublish(&TourVersion{}, []Hint{broken})
		require.Error(t, err)
		assert.GreaterOrEqual(t, len(detailPaths(t, err)), 4)
	})

	t.Run("ошибка публикации это 422", func(t *testing.T) {
		err := validateForPublish(version(), nil)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, 422, ve.StatusCode())

		body := ve.ToHTTPError()
		assert.Equal(t, "Tour cannot be published", body.Message)
		for _, d := range ve.Details {
			assert.Contains(t, body.Error, d.Path)
		}
	})
	t.Run("scroll_depth без конфига не публикуется", func(t *testing.T) {
		v := &TourVersion{TargetPath: "/x", TriggerType: TriggerScrollDepth}
		err := validateForPublish(v, []Hint{validHint(1)})
		require.Error(t, err)
		assert.Contains(t, detailPaths(t, err), "trigger_config.scroll_depth")
	})

	t.Run("scroll_depth с валидным конфигом публикуется", func(t *testing.T) {
		depth := 50
		v := &TourVersion{
			TargetPath:    "/x",
			TriggerType:   TriggerScrollDepth,
			TriggerConfig: &TriggerConfig{ScrollDepth: &depth},
		}
		assert.NoError(t, validateForPublish(v, []Hint{validHint(1)}))
	})
}

func TestMatchAudience(t *testing.T) {
	tests := []struct {
		name     string
		audience Audience
		props    map[string]any
		progress *Progress
		want     bool
	}{
		{"без ограничений показываем", Audience{}, nil, nil, true},
		{
			"only_new пропускает нового",
			Audience{OnlyNew: true}, map[string]any{"isNewUser": true}, nil, true,
		},
		{
			"only_new отсекает старого",
			Audience{OnlyNew: true}, map[string]any{"isNewUser": false}, nil, false,
		},
		{
			"only_new отсекает при отсутствии признака",
			Audience{OnlyNew: true}, nil, nil, false,
		},
		{
			"show_once отсекает завершивших",
			Audience{ShowOnce: true}, nil, &Progress{Status: ProgressCompleted}, false,
		},
		{
			"show_once отсекает закрывших крестиком",
			Audience{ShowOnce: true}, nil, &Progress{Status: ProgressDismissed}, false,
		},
		{
			"show_once не мешает незакончившим",
			Audience{ShowOnce: true}, nil, &Progress{Status: ProgressInProgress}, true,
		},
		{
			"max_shows отсекает по счётчику",
			Audience{MaxShows: 3}, nil, &Progress{Status: ProgressInProgress, ShowsCount: 3}, false,
		},
		{
			"max_shows пропускает до лимита",
			Audience{MaxShows: 3}, nil, &Progress{Status: ProgressInProgress, ShowsCount: 2}, true,
		},
		{
			"max_shows=0 это без ограничения",
			Audience{}, nil, &Progress{Status: ProgressInProgress, ShowsCount: 99}, true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, matchAudience(tt.audience, tt.props, tt.progress))
		})
	}
}

func TestValidateEvent(t *testing.T) {
	hintId := uuid.New()
	versionId := uuid.New()

	t.Run("событие уровня тура без hint_id", func(t *testing.T) {
		assert.NoError(t, validateEvent(&Event{
			EventKey: "k", Type: EventTourStarted, TourVersionId: versionId,
		}))
	})

	t.Run("событие уровня шага с hint_id", func(t *testing.T) {
		assert.NoError(t, validateEvent(&Event{
			EventKey: "k", Type: EventHintShown, TourVersionId: versionId, HintId: &hintId,
		}))
	})

	t.Run("событие уровня тура с hint_id отвергается", func(t *testing.T) {
		assert.Error(t, validateEvent(&Event{
			EventKey: "k", Type: EventTourStarted, TourVersionId: versionId, HintId: &hintId,
		}))
	})

	t.Run("событие уровня шага без hint_id отвергается", func(t *testing.T) {
		assert.Error(t, validateEvent(&Event{
			EventKey: "k", Type: EventHintShown, TourVersionId: versionId,
		}))
	})

	t.Run("EventCustom без hint_id проходит", func(t *testing.T) {
		assert.NoError(t, validateEvent(&Event{
			EventKey: "k", Type: EventCustom, TourVersionId: versionId,
		}))
	})

	t.Run("EventCustom с hint_id тоже проходит (гибкость)", func(t *testing.T) {
		assert.NoError(t, validateEvent(&Event{
			EventKey: "k", Type: EventCustom, TourVersionId: versionId, HintId: &hintId,
		}))
	})

	t.Run("событие без версии отвергается", func(t *testing.T) {
		assert.Error(t, validateEvent(&Event{EventKey: "k", Type: EventTourStarted}))
	})

	t.Run("неизвестный тип отвергается", func(t *testing.T) {
		assert.Error(t, validateEvent(&Event{
			EventKey: "k", Type: EventType("nope"), TourVersionId: versionId,
		}))
	})

	t.Run("событие без ключа отвергается", func(t *testing.T) {
		assert.Error(t, validateEvent(&Event{Type: EventTourStarted, TourVersionId: versionId}))
	})
	t.Run("app-level custom с tour_id отвергается", func(t *testing.T) {
		assert.ErrorIs(t, validateEvent(&Event{
			EventKey: "k", Type: EventCustom,
			TourId: uuid.New(), // ← не должно быть
		}), ErrTourUnexpected)
	})

	t.Run("app-level custom с hint_id отвергается", func(t *testing.T) {
		hintId := uuid.New()
		assert.ErrorIs(t, validateEvent(&Event{
			EventKey: "k", Type: EventCustom,
			HintId: &hintId, // ← не должно быть
		}), ErrHintUnexpected)
	})

	t.Run("app-level custom без привязок проходит", func(t *testing.T) {
		assert.NoError(t, validateEvent(&Event{
			EventKey: "k", Type: EventCustom,
			// TourId, TourVersionId, HintId — все пустые
		}))
	})
	t.Run("tour-bound custom с tour_version_id и hint_id проходит", func(t *testing.T) {
		hintId := uuid.New()
		versionId := uuid.New()
		assert.NoError(t, validateEvent(&Event{
			EventKey: "k", Type: EventCustom,
			TourVersionId: versionId,
			HintId:        &hintId, // разрешено для tour-bound custom
		}))
	})
}

func TestValidateEventScope(t *testing.T) {
	var (
		appA      = uuid.New()
		appB      = uuid.New()
		tourId    = uuid.New()
		hintId    = uuid.New()
		otherHint = uuid.New()
	)

	scope := func() *eventScope {
		return &eventScope{
			TourId: tourId,
			AppId:  appA,
			Hints:  map[uuid.UUID]struct{}{hintId: {}},
		}
	}

	t.Run("своё приложение, свой тур, своя подсказка", func(t *testing.T) {
		assert.NoError(t, validateEventScope(
			&Event{TourId: tourId, HintId: &hintId}, appA, scope()))
	})

	t.Run("чужое приложение отвергается", func(t *testing.T) {
		err := validateEventScope(&Event{TourId: tourId}, appB, scope())
		assert.ErrorIs(t, err, ErrVersionForeignApp)
	})

	t.Run("пустой tour_id отвергается", func(t *testing.T) {
		err := validateEventScope(&Event{TourId: uuid.Nil}, appA, scope())
		assert.ErrorIs(t, err, ErrTourRequired)
	})

	t.Run("версия из другого тура отвергается", func(t *testing.T) {
		err := validateEventScope(&Event{TourId: uuid.New()}, appA, scope())
		assert.ErrorIs(t, err, ErrTourVersionMismatch)
	})

	t.Run("подсказка не из этой версии отвергается", func(t *testing.T) {
		err := validateEventScope(&Event{TourId: tourId, HintId: &otherHint}, appA, scope())
		assert.ErrorIs(t, err, ErrHintNotInVersion)
	})

	t.Run("событие уровня тура без подсказки проходит", func(t *testing.T) {
		assert.NoError(t, validateEventScope(&Event{TourId: tourId}, appA, scope()))
	})
}

func TestEventTypeTourLevel(t *testing.T) {
	assert.True(t, EventGoalReached.TourLevel())
	assert.True(t, EventTourCompleted.TourLevel())
	assert.NotEqual(t, EventGoalReached, EventTourCompleted)

	assert.False(t, EventHintShown.TourLevel())
	assert.False(t, EventSelectorMissing.TourLevel())

	assert.True(t, EventCustom.TourLevel())
	assert.True(t, EventCustom.Valid())
}

func TestTourStatusValid(t *testing.T) {
	assert.True(t, TourDraft.Valid())
	assert.True(t, TourPublished.Valid())
	assert.True(t, TourArchived.Valid())
	assert.False(t, TourStatus("").Valid())
	assert.False(t, TourStatus("live").Valid())
}

func detailPaths(t *testing.T, err error) []string {
	t.Helper()

	var ve *ValidationError
	require.ErrorAs(t, err, &ve)

	paths := make([]string, 0, len(ve.Details))
	for _, d := range ve.Details {
		paths = append(paths, d.Path)
	}
	return paths
}

func TestValidateTriggerConfig(t *testing.T) {

	t.Run("delay без конфига возвращает ошибку", func(t *testing.T) {
		details := validateTriggerConfig(TriggerDelay, nil)
		require.Len(t, details, 1)
		assert.Equal(t, "trigger_config.delay_ms", details[0].Path)
	})

	t.Run("delay с пустым конфигом возвращает ошибку", func(t *testing.T) {
		details := validateTriggerConfig(TriggerDelay, &TriggerConfig{})
		require.Len(t, details, 1)
		assert.Equal(t, "trigger_config.delay_ms", details[0].Path)
	})

	t.Run("delay_ms меньше 100 — ошибка", func(t *testing.T) {
		ms := 50
		details := validateTriggerConfig(TriggerDelay, &TriggerConfig{DelayMs: &ms})
		require.Len(t, details, 1)
	})

	t.Run("delay_ms валидный", func(t *testing.T) {
		ms := 1500
		details := validateTriggerConfig(TriggerDelay, &TriggerConfig{DelayMs: &ms})
		assert.Empty(t, details)
	})

	t.Run("delay_ms больше 600000 — ошибка", func(t *testing.T) {
		ms := 700000
		details := validateTriggerConfig(TriggerDelay, &TriggerConfig{DelayMs: &ms})
		require.Len(t, details, 1)
	})

	t.Run("scroll_depth без конфига возвращает ошибку", func(t *testing.T) {
		details := validateTriggerConfig(TriggerScrollDepth, nil)
		assert.Len(t, details, 1)
		assert.Equal(t, "trigger_config.scroll_depth", details[0].Path)
	})

	t.Run("on_load без конфига — ок", func(t *testing.T) {
		details := validateTriggerConfig(TriggerOnLoad, nil)
		assert.Empty(t, details)
	})

	t.Run("scroll_depth валидный", func(t *testing.T) {
		depth := 50
		details := validateTriggerConfig(TriggerScrollDepth, &TriggerConfig{ScrollDepth: &depth})
		assert.Empty(t, details)
	})

	t.Run("scroll_depth вне диапазона 1-100", func(t *testing.T) {
		depth := 150
		details := validateTriggerConfig(TriggerScrollDepth, &TriggerConfig{ScrollDepth: &depth})
		assert.Len(t, details, 1)
		assert.Equal(t, "trigger_config.scroll_depth", details[0].Path)
	})

	t.Run("scroll_depth равен 0 — ошибка", func(t *testing.T) {
		depth := 0
		details := validateTriggerConfig(TriggerScrollDepth, &TriggerConfig{ScrollDepth: &depth})
		assert.Len(t, details, 1)
	})

	t.Run("inactivity_secs меньше 3 — ошибка", func(t *testing.T) {
		secs := 2
		details := validateTriggerConfig(TriggerInactivity, &TriggerConfig{InactivitySecs: &secs})
		assert.Len(t, details, 1)
		assert.Equal(t, "trigger_config.inactivity_secs", details[0].Path)
	})

	t.Run("inactivity_secs валидный", func(t *testing.T) {
		secs := 10
		details := validateTriggerConfig(TriggerInactivity, &TriggerConfig{InactivitySecs: &secs})
		assert.Empty(t, details)
	})

	t.Run("element_selector пустая строка — ошибка", func(t *testing.T) {
		sel := "   "
		details := validateTriggerConfig(TriggerElementVisible, &TriggerConfig{ElementSelector: &sel})
		assert.Len(t, details, 1)
		assert.Equal(t, "trigger_config.element_selector", details[0].Path)
	})

	t.Run("element_selector валидный", func(t *testing.T) {
		sel := "#checkout-btn"
		details := validateTriggerConfig(TriggerElementVisible, &TriggerConfig{ElementSelector: &sel})
		assert.Empty(t, details)
	})

	t.Run("inactivity без конфига возвращает ошибку", func(t *testing.T) {
		details := validateTriggerConfig(TriggerInactivity, nil)
		require.Len(t, details, 1)
		assert.Equal(t, "trigger_config.inactivity_secs", details[0].Path)
		assert.Contains(t, details[0].Message, "is required")
	})

	t.Run("inactivity с пустым конфигом возвращает ошибку", func(t *testing.T) {
		details := validateTriggerConfig(TriggerInactivity, &TriggerConfig{})
		require.Len(t, details, 1)
		assert.Equal(t, "trigger_config.inactivity_secs", details[0].Path)
	})

	t.Run("element_visible без конфига возвращает ошибку", func(t *testing.T) {
		details := validateTriggerConfig(TriggerElementVisible, nil)
		require.Len(t, details, 1)
		assert.Equal(t, "trigger_config.element_selector", details[0].Path)
		assert.Contains(t, details[0].Message, "is required")
	})

	t.Run("element_visible с пустым конфигом возвращает ошибку", func(t *testing.T) {
		details := validateTriggerConfig(TriggerElementVisible, &TriggerConfig{})
		require.Len(t, details, 1)
		assert.Equal(t, "trigger_config.element_selector", details[0].Path)
	})

	t.Run("element_visible с пустой строкой возвращает ошибку", func(t *testing.T) {
		empty := "   "
		details := validateTriggerConfig(TriggerElementVisible, &TriggerConfig{ElementSelector: &empty})
		require.Len(t, details, 1)
		assert.Equal(t, "trigger_config.element_selector", details[0].Path)
	})

	t.Run("scroll_depth с пустым конфигом возвращает ошибку", func(t *testing.T) {
		details := validateTriggerConfig(TriggerScrollDepth, &TriggerConfig{})
		require.Len(t, details, 1)
		assert.Equal(t, "trigger_config.scroll_depth", details[0].Path)
	})

	for _, tt := range []TriggerType{TriggerOnLoad, TriggerExitIntent, TriggerManual} {
		t.Run(string(tt)+" без конфига — ок", func(t *testing.T) {
			details := validateTriggerConfig(tt, nil)
			assert.Empty(t, details)
		})
	}
}

func TestMatchAudienceRules_Integration(t *testing.T) {
	history := []Event{
		{
			Type:       EventCustom,
			Payload:    map[string]any{"event_name": "checkout_started"},
			OccurredAt: time.Now().Add(-1 * time.Hour),
		},
	}

	t.Run("event_performed с неподходящим оператором не совпадает", func(t *testing.T) {
		rules := []AudienceRule{
			{Type: "event_performed", Key: "checkout_started", Operator: "neq", Timeframe: "24h"},
		}
		assert.False(t, matchAudienceRules(rules, history, nil))
	})

	t.Run("page_visited с contains не совпадает", func(t *testing.T) {
		rules := []AudienceRule{
			{Type: "page_visited", Key: "/pricing", Operator: "contains", Timeframe: "24h"},
		}
		assert.False(t, matchAudienceRules(rules, history, nil))
	})

	t.Run("event_performed not_exists с существующим событием", func(t *testing.T) {
		rules := []AudienceRule{
			{Type: "event_performed", Key: "checkout_started", Operator: "not_exists", Timeframe: "24h"},
		}
		assert.False(t, matchAudienceRules(rules, history, nil))
	})

	t.Run("event_performed not_exists с отсутствующим событием", func(t *testing.T) {
		rules := []AudienceRule{
			{Type: "event_performed", Key: "nonexistent_event", Operator: "not_exists", Timeframe: "24h"},
		}
		assert.True(t, matchAudienceRules(rules, history, nil))
	})

	t.Run("неизвестный rule.Type с not_exists не совпадает", func(t *testing.T) {
		rules := []AudienceRule{
			{Type: "bogus_rule_type", Key: "x", Operator: "not_exists"},
		}
		assert.False(t, matchAudienceRules(rules, history, nil))
	})

	t.Run("пустой rule.Type не совпадает", func(t *testing.T) {
		rules := []AudienceRule{
			{Type: "", Key: "x", Operator: "exists"},
		}
		assert.False(t, matchAudienceRules(rules, history, nil))
	})
}

func TestEvaluateOperator(t *testing.T) {
	t.Run("eq со строками", func(t *testing.T) {
		assert.True(t, evaluateOperator("eq", "hello", "hello"))
		assert.False(t, evaluateOperator("eq", "hello", "world"))
	})

	t.Run("neq", func(t *testing.T) {
		assert.True(t, evaluateOperator("neq", "a", "b"))
		assert.False(t, evaluateOperator("neq", "a", "a"))
	})

	t.Run("числовые операторы", func(t *testing.T) {
		assert.True(t, evaluateOperator("gt", 10, 5))
		assert.True(t, evaluateOperator("gte", 10, 10))
		assert.True(t, evaluateOperator("lt", 5, 10))
		assert.True(t, evaluateOperator("lte", 10, 10))
		assert.False(t, evaluateOperator("gt", 5, 10))
	})

	t.Run("строковые операторы", func(t *testing.T) {
		assert.True(t, evaluateOperator("contains", "hello world", "world"))
		assert.False(t, evaluateOperator("contains", "hello", "world"))
		assert.True(t, evaluateOperator("starts_with", "hello", "hel"))
		assert.True(t, evaluateOperator("ends_with", "hello", "llo"))
	})

	t.Run("nil значения", func(t *testing.T) {
		assert.True(t, evaluateOperator("eq", nil, nil))
		assert.True(t, evaluateOperator("neq", nil, "value"))
		assert.True(t, evaluateOperator("neq", "value", nil))
	})
}

func TestToFloat64(t *testing.T) {
	t.Run("разные числовые типы", func(t *testing.T) {
		v, ok := toFloat64(float64(3.14))
		assert.True(t, ok)
		assert.Equal(t, 3.14, v)

		v, ok = toFloat64(42)
		assert.True(t, ok)
		assert.Equal(t, float64(42), v)

		v, ok = toFloat64(int64(100))
		assert.True(t, ok)
		assert.Equal(t, float64(100), v)
	})

	t.Run("строка-число", func(t *testing.T) {
		v, ok := toFloat64("3.14")
		assert.True(t, ok)
		assert.Equal(t, 3.14, v)
	})

	t.Run("не-число", func(t *testing.T) {
		_, ok := toFloat64("hello")
		assert.False(t, ok)

		_, ok = toFloat64([]int{1, 2, 3})
		assert.False(t, ok)
	})
}

func TestParseTimeframe(t *testing.T) {
	t.Run("пустая строка — безлимитно", func(t *testing.T) {
		d, ok := parseTimeframe("")
		assert.True(t, ok)
		assert.Equal(t, time.Duration(0), d)
	})

	t.Run("all_time — безлимитно", func(t *testing.T) {
		d, ok := parseTimeframe("all_time")
		assert.True(t, ok)
		assert.Equal(t, time.Duration(0), d)
	})

	t.Run("7d — 168 часов", func(t *testing.T) {
		d, ok := parseTimeframe("7d")
		assert.True(t, ok)
		assert.Equal(t, 7*24*time.Hour, d)
	})

	t.Run("90d — невалидно", func(t *testing.T) {
		_, ok := parseTimeframe("90d")
		assert.False(t, ok)
	})

	t.Run("опечатка — невалидно", func(t *testing.T) {
		_, ok := parseTimeframe("7days")
		assert.False(t, ok)
	})
}

func TestTimeframeConsistency(t *testing.T) {
	for _, tf := range []string{"1h", "24h", "7d", "30d", "all_time", "", "90d", "typo"} {
		t.Run(tf, func(t *testing.T) {
			rules := []AudienceRule{{Type: "event_performed", Timeframe: tf}}
			minSince := computeMinSince(rules)

			recentEvent := time.Now().Add(-59 * time.Minute)
			matches := checkTimeframe(recentEvent, tf)

			_, ok := parseTimeframe(tf)
			if !ok {
				assert.True(t, minSince.IsZero(), "invalid timeframe should expand window to unbounded")
				assert.False(t, matches, "invalid timeframe should fail-closed in checkTimeframe")
			} else {
				if minSince.IsZero() {
					assert.True(t, matches, "unbounded window must accept any event")
				} else {
					assert.True(t, matches, "59-min-old event must fit in any valid timeframe ≥ 1h")
				}
			}
		})
	}
}

func TestValidateAudienceRules(t *testing.T) {
	t.Run("пустой список — ок", func(t *testing.T) {
		assert.Empty(t, validateAudienceRules(nil))
	})

	t.Run("валидное правило event_performed", func(t *testing.T) {
		rules := []AudienceRule{{
			Type: "event_performed", Key: "checkout",
			Operator: "exists", Timeframe: "7d",
		}}
		assert.Empty(t, validateAudienceRules(rules))
	})

	t.Run("опечатка в type ловится", func(t *testing.T) {
		rules := []AudienceRule{{
			Type: "pag_visited", Key: "/pricing", // опечатка!
			Operator: "exists", Timeframe: "7d",
		}}
		details := validateAudienceRules(rules)
		require.Len(t, details, 1)
		assert.Contains(t, details[0].Path, "type")
		assert.Contains(t, details[0].Message, "unknown type")
	})

	t.Run("невалидный timeframe ловится", func(t *testing.T) {
		rules := []AudienceRule{{
			Type: "event_performed", Key: "x",
			Operator: "exists", Timeframe: "90d",
		}}
		details := validateAudienceRules(rules)
		require.Len(t, details, 1)
		assert.Contains(t, details[0].Path, "timeframe")
	})

	t.Run("neq для event_performed отвергается", func(t *testing.T) {
		rules := []AudienceRule{{
			Type: "event_performed", Key: "x",
			Operator: "neq", Timeframe: "7d",
		}}
		details := validateAudienceRules(rules)
		require.Len(t, details, 1)
		assert.Contains(t, details[0].Message, "must be exists or not_exists")
	})

	t.Run("user_property не требует timeframe", func(t *testing.T) {
		rules := []AudienceRule{{
			Type: "user_property", Key: "plan",
			Operator: "eq", Value: "pro",
		}}
		assert.Empty(t, validateAudienceRules(rules))
	})

	t.Run("user_property с timeframe отвергается", func(t *testing.T) {
		rules := []AudienceRule{{
			Type: "user_property", Key: "plan",
			Operator: "eq", Value: "pro", Timeframe: "7d",
		}}
		details := validateAudienceRules(rules)
		require.Len(t, details, 1)
		assert.Contains(t, details[0].Path, "timeframe")
	})
}

func TestValidateForPublish_RejectsBadAudienceRules(t *testing.T) {
	validHint := Hint{Step: 1, Title: "t", Content: "c", Selector: "#id", Placement: PlacementBottom}
	v := &TourVersion{
		TargetPath:  "/x",
		TriggerType: TriggerOnLoad,
		Audience: Audience{
			Rules: []AudienceRule{{
				Type: "pag_visited", Key: "/pricing", // опечатка
				Operator: "exists", Timeframe: "7d",
			}},
		},
	}
	err := validateForPublish(v, []Hint{validHint})
	require.Error(t, err)
	paths := detailPaths(t, err)
	assert.Contains(t, paths, "audience.rules[0].type")
}

// Тело черновика валидируется одинаково на всех входах, поэтому проверяем не
// только сам набор правил, но и то, что Create и UpdateDraft его вызывают:
// раньше создание тура принимало конфиг, который потом не публиковался.
type stubTourRepo struct {
	TourRepositoryInterface
	draft   *TourVersion
	created bool
	updated bool
}

func (s *stubTourRepo) CreateWithDraft(_ context.Context, t *Tour) (*Tour, error) {
	s.created = true
	return t, nil
}

func (s *stubTourRepo) VersionByStatus(_ context.Context, _ uuid.UUID, _ TourStatus) (*TourVersion, error) {
	return s.draft, nil
}

func (s *stubTourRepo) UpdateDraft(_ context.Context, _ uuid.UUID, _ DraftPatch) (*TourVersion, error) {
	s.updated = true
	return s.draft, nil
}

func TestValidateDraftBody(t *testing.T) {
	t.Run("собирает проблемы и триггера, и правил", func(t *testing.T) {
		details := validateDraftBody(
			TriggerScrollDepth,
			nil,
			[]AudienceRule{{Type: "user_property", Key: "plan", Operator: "bogus"}},
		)
		paths := make([]string, 0, len(details))
		for _, d := range details {
			paths = append(paths, d.Path)
		}
		assert.Contains(t, paths, "trigger_config.scroll_depth")
		assert.Contains(t, paths, "audience.rules[0].operator")
	})
}

func TestCreateValidatesDraftBody(t *testing.T) {
	appId := uuid.New()

	t.Run("тур с невалидным конфигом не доходит до репозитория", func(t *testing.T) {
		repo := &stubTourRepo{}
		td := NewTourDomain(repo, nil)

		_, err := td.Create(context.Background(), appId, &Tour{
			Title: "t", TargetPath: "/x", TriggerType: TriggerScrollDepth,
		})

		require.Error(t, err)
		assert.False(t, repo.created, "невалидный тур не должен создаваться")
		assert.Contains(t, detailPaths(t, err), "trigger_config.scroll_depth")
	})

	t.Run("тур с невалидным правилом аудитории отвергается", func(t *testing.T) {
		repo := &stubTourRepo{}
		td := NewTourDomain(repo, nil)

		_, err := td.Create(context.Background(), appId, &Tour{
			Title: "t", TargetPath: "/x",
			Audience: &Audience{Rules: []AudienceRule{{
				Type: "event_performed", Key: "checkout", Operator: "neq", Timeframe: "7d",
			}}},
		})

		require.Error(t, err)
		assert.False(t, repo.created)
		assert.Contains(t, detailPaths(t, err), "audience.rules[0].operator")
	})

	t.Run("валидный тур создаётся", func(t *testing.T) {
		repo := &stubTourRepo{}
		td := NewTourDomain(repo, nil)

		depth := 50
		_, err := td.Create(context.Background(), appId, &Tour{
			Title: "t", TargetPath: "/x", TriggerType: TriggerScrollDepth,
			TriggerConfig: &TriggerConfig{ScrollDepth: &depth},
		})

		require.NoError(t, err)
		assert.True(t, repo.created)
	})
}

func TestUpdateDraftValidatesAudienceRules(t *testing.T) {
	draft := &TourVersion{
		Id: uuid.New(), Status: TourDraft,
		TriggerType: TriggerOnLoad, TargetPath: "/x",
	}

	t.Run("невалидное правило в патче отвергается", func(t *testing.T) {
		repo := &stubTourRepo{draft: draft}
		td := NewTourDomain(repo, nil)

		_, err := td.UpdateDraft(context.Background(), uuid.New(), DraftPatch{
			Audience: &Audience{Rules: []AudienceRule{{
				Type: "user_property", Key: "plan", Operator: "eq", Timeframe: "7d",
			}}},
		})

		require.Error(t, err)
		assert.False(t, repo.updated, "невалидный патч не должен доходить до репозитория")
		assert.Contains(t, detailPaths(t, err), "audience.rules[0].timeframe")
	})

	t.Run("валидный патч проходит", func(t *testing.T) {
		repo := &stubTourRepo{draft: draft}
		td := NewTourDomain(repo, nil)

		_, err := td.UpdateDraft(context.Background(), uuid.New(), DraftPatch{
			Audience: &Audience{Rules: []AudienceRule{{
				Type: "user_property", Key: "plan", Operator: "eq", Value: "free",
			}}},
		})

		require.NoError(t, err)
		assert.True(t, repo.updated)
	})
}
