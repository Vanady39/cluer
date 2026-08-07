package repositories

import (
	"context"

	"github.com/Vanady39/cluer/internal/domains"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HintRepository struct {
	pool *pgxpool.Pool
}

func NewHintRepository(pool *pgxpool.Pool) *HintRepository {
	return &HintRepository{pool: pool}
}

const hintColumns = `id, tour_version_id, step, title, content, selector, placement,
	media_url, spotlight, required, wait_for_selector, input_placeholder,
	expected_input, created_at, updated_at`

func (hr *HintRepository) Create(ctx context.Context, h *domains.Hint) error {
	err := hr.pool.QueryRow(ctx, `
		INSERT INTO hints (
			tour_version_id, step, title, content, selector, placement, media_url,
			spotlight, required, wait_for_selector, input_placeholder, expected_input)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id, created_at, updated_at`,
		h.TourVersionId, h.Step, h.Title, h.Content, h.Selector, h.Placement, h.MediaUrl,
		h.Spotlight, h.Required, h.WaitForSelector, h.InputPlaceHolder, h.ExpectedInput,
	).Scan(&h.Id, &h.CreatedAt, &h.UpdatedAt)

	return wrap(err, domains.ErrHintNotFound, h.TourVersionId.String(), domains.Create)
}

func (hr *HintRepository) GetById(ctx context.Context, id uuid.UUID) (*domains.Hint, error) {
	h, err := scanHint(hr.pool.QueryRow(ctx, `SELECT `+hintColumns+` FROM hints WHERE id = $1`, id))
	if err != nil {
		return nil, wrap(err, domains.ErrHintNotFound, id.String(), domains.Retrieve)
	}
	return h, nil
}

func (hr *HintRepository) ListByVersion(ctx context.Context, versionId uuid.UUID) ([]domains.Hint, error) {
	rows, err := hr.pool.Query(ctx,
		`SELECT `+hintColumns+` FROM hints WHERE tour_version_id = $1 ORDER BY step`, versionId)
	if err != nil {
		return nil, wrap(err, domains.ErrHintNotFound, versionId.String(), domains.Retrieve)
	}
	defer rows.Close()

	hints := make([]domains.Hint, 0)
	for rows.Next() {
		h, err := scanHint(rows)
		if err != nil {
			return nil, wrap(err, domains.ErrHintNotFound, versionId.String(), domains.Retrieve)
		}
		hints = append(hints, *h)
	}
	return hints, rows.Err()
}

func (hr *HintRepository) Update(ctx context.Context, h *domains.Hint) error {
	updated, err := scanHint(hr.pool.QueryRow(ctx, `
		UPDATE hints SET title = $2, content = $3, selector = $4, placement = $5,
			media_url = $6, spotlight = $7, required = $8, wait_for_selector = $9,
			input_placeholder = $10, expected_input = $11
		WHERE id = $1
		RETURNING `+hintColumns,
		h.Id, h.Title, h.Content, h.Selector, h.Placement, h.MediaUrl,
		h.Spotlight, h.Required, h.WaitForSelector, h.InputPlaceHolder, h.ExpectedInput))
	if err != nil {
		return wrap(err, domains.ErrHintNotFound, h.Id.String(), domains.Update)
	}
	*h = *updated
	return nil
}

// Delete removes the hint and closes the hole it leaves in the numbering.
//
// Renumbering is not cosmetic. Publishing rejects gaps, so a delete that left
// one behind would make the tour unpublishable with an error pointing at a hint
// the admin never touched.
func (hr *HintRepository) Delete(ctx context.Context, versionId, id uuid.UUID) error {
	tx, err := hr.pool.Begin(ctx)
	if err != nil {
		return wrap(err, domains.ErrHintNotFound, id.String(), domains.Delete)
	}
	defer tx.Rollback(ctx)

	var step int
	err = tx.QueryRow(ctx,
		`DELETE FROM hints WHERE id = $1 AND tour_version_id = $2 RETURNING step`, id, versionId,
	).Scan(&step)
	if err != nil {
		return wrap(err, domains.ErrHintNotFound, id.String(), domains.Delete)
	}

	if _, err := tx.Exec(ctx,
		`UPDATE hints SET step = step - 1 WHERE tour_version_id = $1 AND step > $2`,
		versionId, step); err != nil {
		return wrap(err, domains.ErrHintNotFound, versionId.String(), domains.Update)
	}

	return wrap(tx.Commit(ctx), domains.ErrHintNotFound, id.String(), domains.Delete)
}

// Reorder assigns new step numbers from a full ordered list.
//
// Every id is resolved before anything is written, so a wrong id cannot leave
// the tour half-renumbered. The unique constraint is deferred for the duration
// because the intermediate states are genuinely invalid: swapping steps 1 and 2
// necessarily passes through a moment where two rows claim the same number.
func (hr *HintRepository) Reorder(ctx context.Context, versionId uuid.UUID, ids []uuid.UUID) error {
	tx, err := hr.pool.Begin(ctx)
	if err != nil {
		return wrap(err, domains.ErrHintNotFound, versionId.String(), domains.Update)
	}
	defer tx.Rollback(ctx)

	var count int
	if err := tx.QueryRow(ctx,
		`SELECT count(*) FROM hints WHERE tour_version_id = $1`, versionId).Scan(&count); err != nil {
		return wrap(err, domains.ErrHintNotFound, versionId.String(), domains.Retrieve)
	}
	// A partial list would silently drop the hints it omits to the end in an
	// arbitrary order, so it is refused rather than guessed at.
	if count != len(ids) {
		return &domains.RepositoryError{
			Err:       domains.ErrReorderMismatch,
			EntityRef: versionId.String(),
			Operation: domains.Update,
			Reason:    domains.InvalidReference,
		}
	}

	if _, err := tx.Exec(ctx, `SET CONSTRAINTS ALL DEFERRED`); err != nil {
		return wrap(err, domains.ErrHintNotFound, versionId.String(), domains.Update)
	}

	for i, id := range ids {
		tag, err := tx.Exec(ctx,
			`UPDATE hints SET step = $1 WHERE id = $2 AND tour_version_id = $3`,
			i+1, id, versionId)
		if err != nil {
			return wrap(err, domains.ErrHintNotFound, id.String(), domains.Update)
		}
		if tag.RowsAffected() == 0 {
			return &domains.RepositoryError{
				Err:       domains.ErrReorderMismatch,
				EntityRef: id.String(),
				Operation: domains.Update,
				Reason:    domains.InvalidReference,
			}
		}
	}

	return wrap(tx.Commit(ctx), domains.ErrHintNotFound, versionId.String(), domains.Update)
}

func scanHint(row scanner) (*domains.Hint, error) {
	h := new(domains.Hint)
	err := row.Scan(
		&h.Id, &h.TourVersionId, &h.Step, &h.Title, &h.Content, &h.Selector, &h.Placement,
		&h.MediaUrl, &h.Spotlight, &h.Required, &h.WaitForSelector, &h.InputPlaceHolder,
		&h.ExpectedInput, &h.CreatedAt, &h.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return h, nil
}
