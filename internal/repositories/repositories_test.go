package repositories_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Vanady39/cluer/internal/domains"
	"github.com/Vanady39/cluer/internal/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The domain layer turns a repository miss into a 404, so the not-found contract
// has to hold: a *RepositoryError with Reason NoRecord, still matching the
// sentinel through errors.Is.
func TestTourRepository_NotFoundContract(t *testing.T) {
	repo := repositories.NewTourRepository()
	ctx := context.Background()
	missing := uuid.Must(uuid.NewV7())

	tour, err := repo.GetById(ctx, missing)
	assert.Nil(t, tour)
	require.Error(t, err)
	assert.ErrorIs(t, err, domains.ErrTourNotFound)

	var repoErr *domains.RepositoryError
	require.True(t, errors.As(err, &repoErr))
	assert.Equal(t, domains.NoRecord, repoErr.Reason)
	assert.Equal(t, 404, repoErr.StatusCode())

	assert.ErrorIs(t, repo.DeleteTour(ctx, missing), domains.ErrTourNotFound)
	assert.ErrorIs(t, repo.SetStatus(ctx, missing, domains.TourPublished), domains.ErrTourNotFound)
	assert.ErrorIs(t, repo.UpdateTour(ctx, &domains.Tour{Id: missing}), domains.ErrTourNotFound)
}

// Stored tours must be copied on the way in and out, otherwise a caller mutating
// the value it got back would silently rewrite the repository's state.
func TestTourRepository_StoresCopies(t *testing.T) {
	repo := repositories.NewTourRepository()
	ctx := context.Background()

	original := &domains.Tour{Id: uuid.Must(uuid.NewV7()), Title: "Тур", TargetPath: "/x"}
	require.NoError(t, repo.CreateTour(ctx, original))

	original.Title = "изменено после сохранения"

	stored, err := repo.GetById(ctx, original.Id)
	require.NoError(t, err)
	assert.Equal(t, "Тур", stored.Title)

	stored.Title = "изменено через возвращённый указатель"

	again, err := repo.GetById(ctx, original.Id)
	require.NoError(t, err)
	assert.Equal(t, "Тур", again.Title)
}

func TestHintRepository_ListByTourIsOrderedByStep(t *testing.T) {
	repo := repositories.NewHintRepository()
	ctx := context.Background()
	tourId := uuid.Must(uuid.NewV7())

	for _, step := range []int{3, 1, 2} {
		require.NoError(t, repo.CreateHint(ctx, &domains.Hint{
			Id: uuid.Must(uuid.NewV7()), TourId: tourId, Step: step,
		}))
	}

	hints, err := repo.ListByTour(ctx, tourId)
	require.NoError(t, err)
	require.Len(t, hints, 3)
	assert.Equal(t, []int{1, 2, 3}, []int{hints[0].Step, hints[1].Step, hints[2].Step})
}

// A reorder naming an id that does not belong to the tour must be rejected without
// renumbering any of the ids that came before it.
func TestHintRepository_ReorderIsAllOrNothing(t *testing.T) {
	repo := repositories.NewHintRepository()
	ctx := context.Background()
	tourId := uuid.Must(uuid.NewV7())

	first := uuid.Must(uuid.NewV7())
	second := uuid.Must(uuid.NewV7())
	require.NoError(t, repo.CreateHint(ctx, &domains.Hint{Id: first, TourId: tourId, Step: 1}))
	require.NoError(t, repo.CreateHint(ctx, &domains.Hint{Id: second, TourId: tourId, Step: 2}))

	stranger := uuid.Must(uuid.NewV7())
	err := repo.ReorderHint(ctx, tourId, []uuid.UUID{second, stranger})
	require.Error(t, err)
	assert.ErrorIs(t, err, domains.ErrReorderMismatch)

	hints, err := repo.ListByTour(ctx, tourId)
	require.NoError(t, err)
	require.Len(t, hints, 2)
	assert.Equal(t, first, hints[0].Id, "steps must be untouched after a rejected reorder")
	assert.Equal(t, second, hints[1].Id)

	require.NoError(t, repo.ReorderHint(ctx, tourId, []uuid.UUID{second, first}))
	hints, err = repo.ListByTour(ctx, tourId)
	require.NoError(t, err)
	assert.Equal(t, second, hints[0].Id)
	assert.Equal(t, first, hints[1].Id)
}
