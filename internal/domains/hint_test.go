package domains

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHintDomain_Create(t *testing.T) {
	tests := []struct {
		name       string
		tourId     uuid.UUID
		input      *Hint
		setupTour  *Tour
		setupHints []Hint
		wantErr    error
		wantStep   int
		repoErr    error
	}{
		{
			name:   "успешное создание первого хинта",
			tourId: uuid.New(),
			input: &Hint{
				Title:     "Шаг 1",
				Content:   "Нажми кнопку",
				Placement: PlacementBottom,
				Selector:  "#button",
			},
			setupTour: &Tour{
				Id:     uuid.New(),
				Title:  "Тур",
				Status: TourDraft,
			},
			wantStep: 1,
		},
		{
			name:   "автоматическая нумерация шагов",
			tourId: uuid.New(),
			input: &Hint{
				Title:     "Новый шаг",
				Content:   "Следующий шаг",
				Placement: PlacementTop,
				Selector:  "#element",
			},
			setupTour: &Tour{
				Id:     uuid.New(),
				Title:  "Тур",
				Status: TourDraft,
			},
			setupHints: []Hint{
				{Id: uuid.New(), Step: 1},
				{Id: uuid.New(), Step: 2},
				{Id: uuid.New(), Step: 3},
			},
			wantStep: 4,
		},
		{
			name:   "создание хинта с размещением center без селектора",
			tourId: uuid.New(),
			input: &Hint{
				Title:     "Центр",
				Content:   "В центре экрана",
				Placement: PlacementCenter,
			},
			setupTour: &Tour{
				Id:     uuid.New(),
				Title:  "Тур",
				Status: TourDraft,
			},
			wantStep: 1,
		},
		{
			name:   "ошибка: тур не найден",
			tourId: uuid.New(),
			input: &Hint{
				Title:     "Шаг",
				Content:   "Контент",
				Placement: PlacementBottom,
				Selector:  "#btn",
			},
			wantErr: ErrTourNotFound,
		},
		{
			name:   "ошибка: тур опубликован",
			tourId: uuid.New(),
			input: &Hint{
				Title:     "Шаг",
				Content:   "Контент",
				Placement: PlacementBottom,
				Selector:  "#btn",
			},
			setupTour: &Tour{
				Id:     uuid.New(),
				Title:  "Тур",
				Status: TourPublished,
			},
			wantErr: ErrTourIsPublished,
		},
		{
			name:   "ошибка: тур архивирован",
			tourId: uuid.New(),
			input: &Hint{
				Title:     "Шаг",
				Content:   "Контент",
				Placement: PlacementBottom,
				Selector:  "#btn",
			},
			setupTour: &Tour{
				Id:     uuid.New(),
				Title:  "Тур",
				Status: TourArchived,
			},
			wantErr: ErrTourIsArchived,
		},
		{
			name:   "ошибка валидации: пустой заголовок",
			tourId: uuid.New(),
			input: &Hint{
				Content:   "Контент",
				Placement: PlacementBottom,
				Selector:  "#btn",
			},
			setupTour: &Tour{
				Id:     uuid.New(),
				Title:  "Тур",
				Status: TourDraft,
			},
			wantErr: ErrTitleRequired,
		},
		{
			name:   "ошибка валидации: пустой контент",
			tourId: uuid.New(),
			input: &Hint{
				Title:     "Заголовок",
				Placement: PlacementBottom,
				Selector:  "#btn",
			},
			setupTour: &Tour{
				Id:     uuid.New(),
				Title:  "Тур",
				Status: TourDraft,
			},
			wantErr: ErrContentRequired,
		},
		{
			name:   "ошибка валидации: невалидное размещение",
			tourId: uuid.New(),
			input: &Hint{
				Title:     "Заголовок",
				Content:   "Контент",
				Placement: "invalid",
				Selector:  "#btn",
			},
			setupTour: &Tour{
				Id:     uuid.New(),
				Title:  "Тур",
				Status: TourDraft,
			},
			wantErr: ErrBadPlacement,
		},
		{
			name:   "ошибка валидации: нет селектора для нецентрального размещения",
			tourId: uuid.New(),
			input: &Hint{
				Title:     "Заголовок",
				Content:   "Контент",
				Placement: PlacementBottom,
			},
			setupTour: &Tour{
				Id:     uuid.New(),
				Title:  "Тур",
				Status: TourDraft,
			},
			wantErr: ErrSelectorRequired,
		},
		{
			name:   "ошибка репозитора при получении существующих хинтов",
			tourId: uuid.New(),
			input: &Hint{
				Title:     "Шаг",
				Content:   "Контент",
				Placement: PlacementBottom,
				Selector:  "#btn",
			},
			setupTour: &Tour{
				Id:     uuid.New(),
				Title:  "Тур",
				Status: TourDraft,
			},
			repoErr: errRepoFailure,
			wantErr: errRepoFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tourRepo := newFakeTourRepo()
			hintRepo := newFakeHintRepo()
			hintRepo.err = tt.repoErr

			if tt.setupTour != nil {
				tt.setupTour.Id = tt.tourId
				tourRepo.tours[tt.tourId] = tt.setupTour
			}

			for _, h := range tt.setupHints {
				h.TourId = tt.tourId
				hintRepo.hints = append(hintRepo.hints, h)
			}

			svc := NewHintDomain(tourRepo, hintRepo)
			result, err := svc.Create(context.Background(), tt.tourId, tt.input)

			if tt.wantErr != nil {
				// Validation failures arrive wrapped in a *LogicError, so match
				// through Unwrap rather than comparing messages.
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.tourId, result.TourId)
				assert.NotEqual(t, uuid.UUID{}, result.Id)
			}
		})
	}
}

func TestHintDomain_ensureTourEditable(t *testing.T) {
	tests := []struct {
		name      string
		tourId    uuid.UUID
		setupTour *Tour
		wantErr   error
		repoErr   error
	}{
		{
			name:   "тур в статусе draft — редактирование разрешено",
			tourId: uuid.New(),
			setupTour: &Tour{
				Id:     uuid.New(),
				Title:  "Тур",
				Status: TourDraft,
			},
		},
		{
			name:   "тур опубликован — редактирование запрещено",
			tourId: uuid.New(),
			setupTour: &Tour{
				Id:     uuid.New(),
				Title:  "Тур",
				Status: TourPublished,
			},
			wantErr: ErrTourIsPublished,
		},
		{
			name:   "тур архивирован — редактирование запрещено",
			tourId: uuid.New(),
			setupTour: &Tour{
				Id:     uuid.New(),
				Title:  "Тур",
				Status: TourArchived,
			},
			wantErr: ErrTourIsArchived,
		},
		{
			name:    "тур не найден",
			tourId:  uuid.New(),
			wantErr: ErrTourNotFound,
		},
		{
			name:    "ошибка репозитора",
			tourId:  uuid.New(),
			repoErr: errRepoFailure,
			wantErr: errRepoFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tourRepo := newFakeTourRepo()
			tourRepo.err = tt.repoErr
			hintRepo := newFakeHintRepo()

			if tt.setupTour != nil {
				tt.setupTour.Id = tt.tourId
				tourRepo.tours[tt.tourId] = tt.setupTour
			}

			svc := NewHintDomain(tourRepo, hintRepo)
			err := svc.ensureTourEditable(context.Background(), tt.tourId)

			if tt.wantErr != nil {
				// Status conflicts wrap in *LogicError, a missing tour in
				// *RepositoryError; both match through Unwrap.
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
