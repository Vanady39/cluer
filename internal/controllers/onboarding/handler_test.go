package onboarding

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/Vanady39/cluer/internal/domains/onboarding"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockTourRepository struct {
	mu           sync.RWMutex
	Tours        map[uuid.UUID]*onboarding.Tour
	CreateErr    error
	GetByIdErr   error
	ListErr      error
	UpdateErr    error
	DeleteErr    error
	SetStatusErr error
	ListPubErr   error
}

func NewMockTourRepository() *MockTourRepository {
	return &MockTourRepository{Tours: make(map[uuid.UUID]*onboarding.Tour)}
}

func (m *MockTourRepository) CreateTour(ctx context.Context, t *onboarding.Tour) error {
	if m.CreateErr != nil {
		return m.CreateErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Tours[t.Id] = t
	return nil
}

func (m *MockTourRepository) GetById(ctx context.Context, id uuid.UUID) (*onboarding.Tour, error) {
	if t, ok := m.Tours[id]; ok {
		return t, nil
	}
	return nil, nil
}

func (m *MockTourRepository) ListTour(ctx context.Context) ([]*onboarding.Tour, error) {
	if m.ListErr != nil {
		return nil, m.ListErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*onboarding.Tour
	for _, t := range m.Tours {
		result = append(result, t)
	}
	return result, nil
}

func (m *MockTourRepository) UpdateTour(ctx context.Context, tour *onboarding.Tour) error { return nil }
func (m *MockTourRepository) DeleteTour(ctx context.Context, id uuid.UUID) error          { return nil }
func (m *MockTourRepository) SetStatus(ctx context.Context, id uuid.UUID, s onboarding.TourStatus) error {
	return nil
}

func (m *MockTourRepository) ListPublished(ctx context.Context, path string) ([]*onboarding.Tour, error) {
	if m.ListPubErr != nil {
		return nil, m.ListPubErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*onboarding.Tour
	for _, t := range m.Tours {
		if t.Status == onboarding.TourPublished && t.TargetPath == path {
			result = append(result, t)
		}
	}
	return result, nil
}

type MockHintRepository struct {
	mu            sync.RWMutex
	Hints         map[uuid.UUID]*onboarding.Hint
	CreateErr     error
	GetByIdErr    error
	ListByTourErr error
	UpdateErr     error
	DeleteErr     error
	ReorderErr    error
}

func NewMockHintRepository() *MockHintRepository {
	return &MockHintRepository{Hints: make(map[uuid.UUID]*onboarding.Hint)}
}

func (m *MockHintRepository) CreateHint(ctx context.Context, h *onboarding.Hint) error {
	if m.CreateErr != nil {
		return m.CreateErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Hints[h.Id] = h
	return nil
}

func (m *MockHintRepository) GetById(ctx context.Context, id uuid.UUID) (*onboarding.Hint, error) {
	if m.GetByIdErr != nil {
		return nil, m.GetByIdErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if h, ok := m.Hints[id]; ok {
		return h, nil
	}
	return nil, nil
}

func (m *MockHintRepository) ListByTour(ctx context.Context, tourId uuid.UUID) ([]onboarding.Hint, error) {
	if m.ListByTourErr != nil {
		return nil, m.ListByTourErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []onboarding.Hint
	for _, h := range m.Hints {
		if h.TourId == tourId {
			result = append(result, *h)
		}
	}
	return result, nil
}

func (m *MockHintRepository) UpdateHint(ctx context.Context, h *onboarding.Hint) error {
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Hints[h.Id] = h
	return nil
}

func (m *MockHintRepository) DeleteHint(ctx context.Context, tourId, id uuid.UUID) error {
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if h, ok := m.Hints[id]; ok && h.TourId == tourId {
		delete(m.Hints, id)
	}
	return nil
}

func (m *MockHintRepository) ReorderHint(ctx context.Context, tourId uuid.UUID, ids []uuid.UUID) error {
	if m.ReorderErr != nil {
		return m.ReorderErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, id := range ids {
		if h, ok := m.Hints[id]; ok && h.TourId == tourId {
			h.Step = i + 1
		}
	}
	return nil
}

type testSetup struct {
	handler  *Handler
	tourRepo *MockTourRepository
	hintRepo *MockHintRepository
}

func setupTestHandler() *testSetup {
	tourRepo := NewMockTourRepository()
	hintRepo := NewMockHintRepository()
	tourService := onboarding.NewTourService(tourRepo, hintRepo)
	hintService := onboarding.NewHintService(tourRepo, hintRepo)
	return &testSetup{
		handler:  NewHandler(tourService, hintService),
		tourRepo: tourRepo,
		hintRepo: hintRepo,
	}
}

func TestHandler_CreateTour_Success(t *testing.T) {
	s := setupTestHandler()

	tour := onboarding.Tour{
		Title:      "My Tour",
		TargetPath: "/home",
		Priority:   1,
	}
	body, _ := json.Marshal(tour)

	req := httptest.NewRequest(http.MethodPost, "/tours", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handler.CreateTour(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var created onboarding.Tour
	err := json.Unmarshal(rr.Body.Bytes(), &created)
	require.NoError(t, err)
	assert.Equal(t, "My Tour", created.Title)
	assert.Equal(t, "/home", created.TargetPath)
	assert.Equal(t, 1, created.Priority)
	assert.Equal(t, onboarding.TourDraft, created.Status)
	assert.NotEqual(t, uuid.UUID{}, created.Id)
}

func TestHandler_CreateTour_ValidationError(t *testing.T) {
	s := setupTestHandler()

	tour := onboarding.Tour{TargetPath: "/home", Priority: 1}
	body, _ := json.Marshal(tour)

	req := httptest.NewRequest(http.MethodPost, "/tours", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	s.handler.CreateTour(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "title is required")
}

func TestHandler_CreateTour_RepoError(t *testing.T) {
	s := setupTestHandler()
	s.tourRepo.CreateErr = errors.New("database connection failed")

	tour := onboarding.Tour{
		Title:      "Test",
		TargetPath: "/test",
		Priority:   1,
	}
	body, _ := json.Marshal(tour)

	req := httptest.NewRequest(http.MethodPost, "/tours", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	s.handler.CreateTour(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "database connection failed")
}

// === ТЕСТЫ ДЛЯ POST /tours/{tourId}/hints ===

func TestHandler_CreateHint_Success(t *testing.T) {
	s := setupTestHandler()

	tour, err := s.handler.TourService.Create(context.Background(), &onboarding.Tour{
		Title: "Tour", TargetPath: "/path", Priority: 1,
	})
	require.NoError(t, err)

	hint := onboarding.Hint{
		Title:     "Step 1",
		Content:   "Click here",
		Placement: onboarding.PlacementBottom,
		Selector:  "#btn",
	}
	body, _ := json.Marshal(hint)

	req := httptest.NewRequest(http.MethodPost, "/tours/"+tour.Id.String()+"/hints", bytes.NewReader(body))
	req.SetPathValue("tourId", tour.Id.String())
	rr := httptest.NewRecorder()

	s.handler.CreateHint(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var created onboarding.Hint
	err = json.Unmarshal(rr.Body.Bytes(), &created)
	require.NoError(t, err)
	assert.Equal(t, "Step 1", created.Title)
	assert.Equal(t, 1, created.Step)
	assert.Equal(t, tour.Id, created.TourId)
}

func TestHandler_CreateHint_InvalidTourId(t *testing.T) {
	s := setupTestHandler()

	hint := onboarding.Hint{Title: "Test", Content: "Test", Placement: onboarding.PlacementCenter}
	body, _ := json.Marshal(hint)

	req := httptest.NewRequest(http.MethodPost, "/tours/not-a-uuid/hints", bytes.NewReader(body))
	req.SetPathValue("tourId", "not-a-uuid")
	rr := httptest.NewRecorder()

	s.handler.CreateHint(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid tour ID")
}

func TestHandler_CreateHint_TourNotFound(t *testing.T) {
	s := setupTestHandler()

	hint := onboarding.Hint{Title: "Test", Content: "Test", Placement: onboarding.PlacementCenter}
	body, _ := json.Marshal(hint)

	fakeId := uuid.New().String()
	req := httptest.NewRequest(http.MethodPost, "/tours/"+fakeId+"/hints", bytes.NewReader(body))
	req.SetPathValue("tourId", fakeId)
	rr := httptest.NewRecorder()

	s.handler.CreateHint(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "tour not found")
}

func TestHandler_CreateHint_TourPublished(t *testing.T) {
	s := setupTestHandler()

	tour, err := s.handler.TourService.Create(context.Background(), &onboarding.Tour{
		Title: "Test", TargetPath: "/path", Priority: 1,
	})
	require.NoError(t, err)

	tour.Status = onboarding.TourPublished
	s.tourRepo.Tours[tour.Id] = tour

	hint := onboarding.Hint{Title: "Test", Content: "Test", Placement: onboarding.PlacementCenter}
	body, _ := json.Marshal(hint)

	req := httptest.NewRequest(http.MethodPost, "/tours/"+tour.Id.String()+"/hints", bytes.NewReader(body))
	req.SetPathValue("tourId", tour.Id.String())
	rr := httptest.NewRecorder()

	s.handler.CreateHint(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Contains(t, rr.Body.String(), "tour is published")
}

func TestHandler_CreateHint_ValidationError(t *testing.T) {
	s := setupTestHandler()

	tour, err := s.handler.TourService.Create(context.Background(), &onboarding.Tour{
		Title: "Tour", TargetPath: "/path", Priority: 1,
	})
	require.NoError(t, err)

	hint := onboarding.Hint{
		Title:     "Test",
		Placement: onboarding.PlacementBottom,
		Selector:  "#btn",
	}
	body, _ := json.Marshal(hint)

	req := httptest.NewRequest(http.MethodPost, "/tours/"+tour.Id.String()+"/hints", bytes.NewReader(body))
	req.SetPathValue("tourId", tour.Id.String())
	rr := httptest.NewRecorder()

	s.handler.CreateHint(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "content is required")
}

func TestHandler_CreateHint_RepoError(t *testing.T) {
	s := setupTestHandler()
	s.hintRepo.CreateErr = errors.New("database error")

	tour, err := s.handler.TourService.Create(context.Background(), &onboarding.Tour{
		Title: "Tour", TargetPath: "/path", Priority: 1,
	})
	require.NoError(t, err)

	hint := onboarding.Hint{
		Title:     "Test",
		Content:   "Test",
		Placement: onboarding.PlacementCenter,
	}
	body, _ := json.Marshal(hint)

	req := httptest.NewRequest(http.MethodPost, "/tours/"+tour.Id.String()+"/hints", bytes.NewReader(body))
	req.SetPathValue("tourId", tour.Id.String())
	rr := httptest.NewRecorder()

	s.handler.CreateHint(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "database error")
}

// === ТЕСТЫ ДЛЯ GET /tours/published ===

func TestHandler_GetPublishedTours_Success(t *testing.T) {
	s := setupTestHandler()

	tourId := uuid.New()
	s.tourRepo.Tours[tourId] = &onboarding.Tour{
		Id:         tourId,
		Title:      "Tour",
		TargetPath: "/dashboard",
		Status:     onboarding.TourPublished,
	}

	req := httptest.NewRequest(http.MethodGet, "/tours/published?path=/dashboard", nil)
	rr := httptest.NewRecorder()

	s.handler.GetPublishedTours(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Tour")
}

func TestHandler_GetPublishedTours_MissingPath(t *testing.T) {
	s := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/tours/published", nil)
	rr := httptest.NewRecorder()

	s.handler.GetPublishedTours(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "path query parameter is required")
}

func TestHandler_GetPublishedTours_RepoError(t *testing.T) {
	s := setupTestHandler()
	s.tourRepo.ListPubErr = errors.New("database error")

	req := httptest.NewRequest(http.MethodGet, "/tours/published?path=/test", nil)
	rr := httptest.NewRecorder()

	s.handler.GetPublishedTours(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "database error")
}
