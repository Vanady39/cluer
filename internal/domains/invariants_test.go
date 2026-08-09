package domains

import (
	"testing"

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
