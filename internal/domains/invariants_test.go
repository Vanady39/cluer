package domains

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The checks below cover the rules the whole design rests on. They are pure
// functions on purpose: the invariants that need a database to break — one
// published version per tour, frozen bodies — are enforced by the schema and
// are verified against a real Postgres, not mocked here.

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
		// Админ, написавший /dashboard, имеет в виду страницу, а не её
		// отсутствие query-параметров.
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
	// Хост намеренно отбрасывается: один и тот же тур должен работать
	// на localhost, стенде и проде без правки паттерна.
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
		// Шаги 1 и 3: SDK остановится на дыре, и хвост тура никто не увидит.
		err := validateForPublish(version(), []Hint{validHint(1), validHint(3)})
		require.Error(t, err)
		assert.Contains(t, detailPaths(t, err), "hints")
	})

	t.Run("дубликат шага ловится", func(t *testing.T) {
		err := validateForPublish(version(), []Hint{validHint(1), validHint(1)})
		require.Error(t, err)
		assert.Contains(t, detailPaths(t, err), "hints[1].step")
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
		// Заставлять админа пересохранять ради следующей ошибки — плохой размен.
		assert.GreaterOrEqual(t, len(detailPaths(t, err)), 4)
	})

	t.Run("ошибка публикации это 422", func(t *testing.T) {
		err := validateForPublish(version(), nil)
		var ve *ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, 422, ve.StatusCode())
		assert.Equal(t, "VALIDATION_FAILED", ve.Code())
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

	t.Run("событие без версии отвергается", func(t *testing.T) {
		// Без версии воронка смешает показы до и после правки текстов.
		assert.Error(t, validateEvent(&Event{EventKey: "k", Type: EventTourStarted}))
	})

	t.Run("неизвестный тип отвергается", func(t *testing.T) {
		assert.Error(t, validateEvent(&Event{
			EventKey: "k", Type: EventType("nope"), TourVersionId: versionId,
		}))
	})

	t.Run("событие без ключа отвергается", func(t *testing.T) {
		// event_key — основа идемпотентности; без него ретрай удвоит метрики.
		assert.Error(t, validateEvent(&Event{Type: EventTourStarted, TourVersionId: versionId}))
	})
}

func TestEventTypeTourLevel(t *testing.T) {
	// goal_reached учитывается отдельно от tour_completed: пользователь может
	// прокликать все подсказки и не опубликовать объявление. Слить их значит
	// заставить метрику измерять саму себя.
	assert.True(t, EventGoalReached.TourLevel())
	assert.True(t, EventTourCompleted.TourLevel())
	assert.NotEqual(t, EventGoalReached, EventTourCompleted)

	assert.False(t, EventHintShown.TourLevel())
	assert.False(t, EventSelectorMissing.TourLevel())
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
