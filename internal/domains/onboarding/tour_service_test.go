package onboarding

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTourRepo struct {
	tours map[uuid.UUID]*Tour
	err   error
}

func newFakeTourRepo() *fakeTourRepo {
	return &fakeTourRepo{tours: make(map[uuid.UUID]*Tour)}
}

func (f *fakeTourRepo) CreateTour(ctx context.Context, t *Tour) error {
	if f.err != nil {
		return f.err
	}
	f.tours[t.Id] = t
	return nil
}

func (f *fakeTourRepo) GetById(ctx context.Context, id uuid.UUID) (*Tour, error) {
	if f.err != nil {
		return nil, f.err
	}
	if t, ok := f.tours[id]; ok {
		return t, nil
	}
	return nil, nil
}

func (f *fakeTourRepo) ListTour(ctx context.Context) ([]*Tour, error) {
	return nil, f.err
}

func (f *fakeTourRepo) UpdateTour(ctx context.Context, t *Tour) error {
	return f.err
}

func (f *fakeTourRepo) DeleteTour(ctx context.Context, id uuid.UUID) error {
	return f.err
}

func (f *fakeTourRepo) SetStatus(ctx context.Context, id uuid.UUID, s TourStatus) error {
	return f.err
}

func (f *fakeTourRepo) ListPublished(ctx context.Context, path string) ([]*Tour, error) {
	if f.err != nil {
		return nil, f.err
	}
	var result []*Tour
	for _, t := range f.tours {
		if t.Status == TourPublished && t.TargetPath == path {
			result = append(result, t)
		}
	}
	return result, nil
}

type fakeHintRepo struct {
	hints []Hint
	err   error
}

func newFakeHintRepo() *fakeHintRepo {
	return &fakeHintRepo{}
}

func (f *fakeHintRepo) CreateHint(ctx context.Context, h *Hint) error {
	if f.err != nil {
		return f.err
	}
	f.hints = append(f.hints, *h)
	return nil
}

func (f *fakeHintRepo) GetById(ctx context.Context, id uuid.UUID) (*Hint, error) {
	return nil, f.err
}

func (f *fakeHintRepo) ListByTour(ctx context.Context, tourId uuid.UUID) ([]Hint, error) {
	if f.err != nil {
		return nil, f.err
	}
	var result []Hint
	for _, h := range f.hints {
		if h.TourId == tourId {
			result = append(result, h)
		}
	}
	return result, nil
}

func (f *fakeHintRepo) UpdateHint(ctx context.Context, h *Hint) error {
	return f.err
}

func (f *fakeHintRepo) DeleteHint(ctx context.Context, tourId, id uuid.UUID) error {
	return f.err
}

func (f *fakeHintRepo) ReorderHint(ctx context.Context, tourId uuid.UUID, ids []uuid.UUID) error {
	return f.err
}

func TestTourService_Create(t *testing.T) {
	tests := []struct {
		name       string
		input      *Tour
		wantErr    error
		repoErr    error
		wantStatus TourStatus
		wantIDSet  bool
	}{
		{
			name: "успешное создание с валидными данными",
			input: &Tour{
				Title:      "Мой тур",
				TargetPath: "/dashboard",
				Priority:   1,
			},
			wantStatus: TourDraft,
			wantIDSet:  true,
		},
		{
			name: "создание с пустым заголовком",
			input: &Tour{
				TargetPath: "/home",
				Priority:   1,
			},
			wantErr: ErrTitleRequired,
		},
		{
			name: "создание с пустым путём",
			input: &Tour{
				Title:    "Тур",
				Priority: 1,
			},
			wantErr: ErrTargetPathRequired,
		},
		{
			name: "создание с отрицательным приоритетом",
			input: &Tour{
				Title:      "Тур",
				TargetPath: "/",
				Priority:   -1,
			},
			wantErr: ErrPriorityNegative,
		},
		{
			name: "создание с невалидным типом триггера",
			input: &Tour{
				Title:       "Тур",
				TargetPath:  "/",
				Priority:    1,
				TriggerType: "invalid_trigger",
			},
			wantErr: ErrInvalidTriggerType,
		},
		{
			name: "создание с валидным типом триггера",
			input: &Tour{
				Title:       "Тур",
				TargetPath:  "/",
				Priority:    1,
				TriggerType: TriggerOnLoad,
			},
			wantStatus: TourDraft,
			wantIDSet:  true,
		},
		{
			name: "создание с пустым типом триггера (допустимо)",
			input: &Tour{
				Title:      "Тур",
				TargetPath: "/",
				Priority:   1,
			},
			wantStatus: TourDraft,
			wantIDSet:  true,
		},
		{
			name: "ошибка репозитория при создании",
			input: &Tour{
				Title:      "Тур",
				TargetPath: "/",
				Priority:   1,
			},
			repoErr: errors.New("database error"),
			wantErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tourRepo := newFakeTourRepo()
			tourRepo.err = tt.repoErr
			hintRepo := newFakeHintRepo()

			svc := NewTourService(tourRepo, hintRepo)
			result, err := svc.Create(context.Background(), tt.input)

			if tt.wantErr != nil {
				assert.Error(t, err)
				if !errors.Is(tt.wantErr, errors.New("")) {
					assert.Equal(t, tt.wantErr.Error(), err.Error())
					assert.Nil(t, result)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.wantStatus, result.Status)

				if tt.wantIDSet {
					assert.NotEqual(t, uuid.UUID{}, result.Id)
					assert.False(t, result.CreatedAt.IsZero())
					assert.False(t, result.UpdatedAt.IsZero())
				}
			}
		})
	}
}

func TestTourService_ListPublished(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		setupTours []*Tour
		setupHints []Hint
		wantCount  int
		repoErr    error
		wantErr    bool
	}{
		{
			name: "возвращает опубликованные туры для пути",
			path: "/dashboard",
			setupTours: []*Tour{
				{
					Id:         uuid.New(),
					Title:      "Тур 1",
					TargetPath: "/dashboard",
					Status:     TourPublished,
				},
				{
					Id:         uuid.New(),
					Title:      "Тур 2",
					TargetPath: "/dashboard",
					Status:     TourPublished,
				},
			},
			wantCount: 2,
		},
		{
			name: "игнорирует неопубликованные туры",
			path: "/dashboard",
			setupTours: []*Tour{
				{
					Id:         uuid.New(),
					Title:      "Черновик",
					TargetPath: "/dashboard",
					Status:     TourDraft,
				},
				{
					Id:         uuid.New(),
					Title:      "Архив",
					TargetPath: "/dashboard",
					Status:     TourArchived,
				},
			},
			wantCount: 0,
		},
		{
			name: "игнорирует туры с другим путём",
			path: "/dashboard",
			setupTours: []*Tour{
				{
					Id:         uuid.New(),
					Title:      "Тур для главной",
					TargetPath: "/home",
					Status:     TourPublished,
				},
			},
			wantCount: 0,
		},
		{
			name: "обогащает туры хинтами",
			path: "/dashboard",
			setupTours: []*Tour{
				{
					Id:         uuid.New(),
					Title:      "Тур с хинтами",
					TargetPath: "/dashboard",
					Status:     TourPublished,
				},
			},
			setupHints: []Hint{
				{
					Id:      uuid.New(),
					TourId:  uuid.Nil,
					Title:   "Шаг 1",
					Content: "Нажми сюда",
					Step:    1,
				},
				{
					Id:      uuid.New(),
					TourId:  uuid.Nil,
					Title:   "Шаг 2",
					Content: "Затем сюда",
					Step:    2,
				},
			},
			wantCount: 1,
		},
		{
			name:    "ошибка репозитора при получении туров",
			path:    "/dashboard",
			repoErr: errors.New("database error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tourRepo := newFakeTourRepo()
			tourRepo.err = tt.repoErr
			hintRepo := newFakeHintRepo()
			hintRepo.err = tt.repoErr

			for _, tour := range tt.setupTours {
				tourRepo.tours[tour.Id] = tour
			}

			if len(tt.setupTours) > 0 && len(tt.setupHints) > 0 {
				for i := range tt.setupHints {
					tt.setupHints[i].TourId = tt.setupTours[0].Id
					hintRepo.hints = append(hintRepo.hints, tt.setupHints[i])
				}
			}

			svc := NewTourService(tourRepo, hintRepo)
			result, err := svc.ListPublished(context.Background(), tt.path)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Len(t, result, tt.wantCount)

				if len(tt.setupHints) > 0 && tt.wantCount > 0 {
					assert.Len(t, result[0].Hints, len(tt.setupHints))
				}
			}
		})
	}
}
