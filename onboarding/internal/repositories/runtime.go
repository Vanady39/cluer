package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Vanady39/cluer/platform/errs"

	"github.com/Vanady39/cluer/onboarding/internal/domains"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type RuntimeRepository struct {
	pool   *pgxpool.Pool
	logger *zerolog.Logger
}

func NewRuntimeRepository(pool *pgxpool.Pool, logger *zerolog.Logger) *RuntimeRepository {
	return &RuntimeRepository{pool: pool, logger: logger}
}

func (rr *RuntimeRepository) Ping(ctx context.Context) error {
	return rr.pool.Ping(ctx)
}

func (rr *RuntimeRepository) CreateApp(ctx context.Context, a *domains.App) error {
	err := rr.pool.QueryRow(ctx, `
		INSERT INTO apps (name, public_key, allowed_origins)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`,
		a.Name, a.PublicKey, a.AllowedOrigins,
	).Scan(&a.Id, &a.CreatedAt)

	return wrap(err, domains.ErrAppNotFound, a.PublicKey, errs.Create)
}

func (rr *RuntimeRepository) AppByKey(ctx context.Context, key string) (*domains.App, error) {
	a := new(domains.App)
	err := rr.pool.QueryRow(ctx, `
		SELECT id, name, public_key, allowed_origins, created_at, archived_at
		FROM apps
		WHERE public_key = $1 AND archived_at IS NULL`, key,
	).Scan(&a.Id, &a.Name, &a.PublicKey, &a.AllowedOrigins, &a.CreatedAt, &a.ArchivedAt)
	if err != nil {
		return nil, wrap(err, domains.ErrAppNotFound, key, errs.Retrieve)
	}
	return a, nil
}

func (rr *RuntimeRepository) UpdateApp(ctx context.Context, id uuid.UUID, name *string, allowedOrigins []string) (*domains.App, error) {
	set := make([]string, 0, 2)
	args := make([]any, 0, 3)

	if name != nil {
		args = append(args, *name)
		set = append(set, fmt.Sprintf("name = $%d", len(args)))
	}
	if allowedOrigins != nil {
		args = append(args, allowedOrigins)
		set = append(set, fmt.Sprintf("allowed_origins = $%d", len(args)))
	}
	if len(set) == 0 {
		return rr.AppByKey(ctx, "")
	}

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE apps SET %s
		WHERE id = $%d AND archived_at IS NULL
		RETURNING id, name, public_key, allowed_origins, created_at, archived_at`,
		strings.Join(set, ", "), len(args))

	a := new(domains.App)
	err := rr.pool.QueryRow(ctx, query, args...).Scan(
		&a.Id, &a.Name, &a.PublicKey, &a.AllowedOrigins, &a.CreatedAt, &a.ArchivedAt)
	if err != nil {
		return nil, wrap(err, domains.ErrAppNotFound, id.String(), errs.Update)
	}
	return a, nil
}

func (rr *RuntimeRepository) ListApps(ctx context.Context) ([]*domains.App, error) {
	rows, err := rr.pool.Query(ctx, `
		SELECT id, name, public_key, allowed_origins, created_at, archived_at
		FROM apps WHERE archived_at IS NULL ORDER BY created_at`)
	if err != nil {
		return nil, wrap(err, domains.ErrAppNotFound, "", errs.Retrieve)
	}
	defer rows.Close()

	apps := make([]*domains.App, 0)
	for rows.Next() {
		a := new(domains.App)
		if err := rows.Scan(&a.Id, &a.Name, &a.PublicKey, &a.AllowedOrigins,
			&a.CreatedAt, &a.ArchivedAt); err != nil {
			return nil, wrap(err, domains.ErrAppNotFound, "", errs.Retrieve)
		}
		apps = append(apps, a)
	}
	return apps, rows.Err()
}

func (rr *RuntimeRepository) SubjectProgress(
	ctx context.Context,
	appId uuid.UUID,
	subjectId string,
) (map[uuid.UUID]*domains.Progress, error) {
	rows, err := rr.pool.Query(ctx, `
		SELECT app_id, subject_id, tour_id, tour_version_id, current_hint_id,
		       status, shows_count, started_at, updated_at, finished_at
		FROM tour_progress
		WHERE app_id = $1 AND subject_id = $2`, appId, subjectId)
	if err != nil {
		return nil, wrap(err, domains.ErrTourNotFound, subjectId, errs.Retrieve)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]*domains.Progress)
	for rows.Next() {
		p := new(domains.Progress)
		if err := rows.Scan(&p.AppId, &p.SubjectId, &p.TourId, &p.TourVersionId,
			&p.CurrentHintId, &p.Status, &p.ShowsCount,
			&p.StartedAt, &p.UpdatedAt, &p.FinishedAt); err != nil {
			return nil, wrap(err, domains.ErrTourNotFound, subjectId, errs.Retrieve)
		}
		result[p.TourId] = p
	}
	return result, rows.Err()
}

func (rr *RuntimeRepository) UpsertProgress(ctx context.Context, p *domains.Progress) error {
	_, err := rr.pool.Exec(ctx, `
		INSERT INTO tour_progress (
			app_id, subject_id, tour_id, tour_version_id, current_hint_id,
			status, shows_count, started_at, finished_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (app_id, subject_id, tour_id) DO UPDATE SET
			tour_version_id = EXCLUDED.tour_version_id,
			current_hint_id = EXCLUDED.current_hint_id,
			status          = EXCLUDED.status,
			shows_count     = EXCLUDED.shows_count,
			started_at      = EXCLUDED.started_at,
			finished_at     = EXCLUDED.finished_at`,
		p.AppId, p.SubjectId, p.TourId, p.TourVersionId, p.CurrentHintId,
		p.Status, p.ShowsCount, p.StartedAt, p.FinishedAt)

	return wrap(err, domains.ErrTourNotFound, p.SubjectId, errs.Update)
}

func (rr *RuntimeRepository) InsertEvents(ctx context.Context, events []domains.Event) (int, int, error) {
	tx, err := rr.pool.Begin(ctx)
	if err != nil {
		return 0, 0, wrap(err, domains.ErrTourNotFound, "", errs.Create)
	}
	defer tx.Rollback(ctx)

	accepted, duplicates := 0, 0

	for i := range events {
		e := events[i]

		var tourID, versionID, hintID any
		if e.TourId != uuid.Nil {
			tourID = e.TourId
		}
		if e.TourVersionId != uuid.Nil {
			versionID = e.TourVersionId
		}
		if e.HintId != nil {
			hintID = *e.HintId
		}

		tag, err := tx.Exec(ctx, `
			INSERT INTO tour_events (
				app_id, event_key, tour_id, tour_version_id, hint_id,
				session_id, subject_id, type, occurred_at, payload)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			ON CONFLICT (app_id, event_key) DO NOTHING`,
			e.AppId, e.EventKey, tourID, versionID, hintID,
			e.SessionId, e.SubjectId, e.Type, e.OccurredAt, e.Payload)
		if err != nil {
			return 0, 0, wrap(err, domains.ErrTourNotFound, e.EventKey, errs.Create)
		}

		if tag.RowsAffected() == 0 {
			duplicates++
			continue
		}
		accepted++

		if e.TourVersionId != uuid.Nil {
			if err := applyProgress(ctx, tx, &e); err != nil {
				return 0, 0, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, 0, wrap(err, domains.ErrTourNotFound, "", errs.Create)
	}
	return accepted, duplicates, nil
}

func applyProgress(ctx context.Context, tx pgx.Tx, e *domains.Event) error {
	switch e.Type {
	case domains.EventHintCompleted:
		_, err := tx.Exec(ctx, `
			UPDATE tour_progress SET current_hint_id = (
				SELECT next.id FROM hints next
				WHERE next.tour_version_id = $4
				  AND next.step > (SELECT step FROM hints WHERE id = $5)
				ORDER BY next.step LIMIT 1)
			WHERE app_id = $1 AND subject_id = $2 AND tour_id = $3
			  AND tour_version_id = $4
			  AND status = 'in_progress'`,
			e.AppId, e.SubjectId, e.TourId, e.TourVersionId, e.HintId)
		return wrap(err, domains.ErrTourNotFound, e.SubjectId, errs.Update)

	case domains.EventTourCompleted, domains.EventTourDismissed:
		status := domains.ProgressCompleted
		if e.Type == domains.EventTourDismissed {
			status = domains.ProgressDismissed
		}
		_, err := tx.Exec(ctx, `
			UPDATE tour_progress
			SET status = $5, finished_at = now(), current_hint_id = NULL
			WHERE app_id = $1 AND subject_id = $2 AND tour_id = $3
			  AND tour_version_id = $4
			  AND status = 'in_progress'`,
			e.AppId, e.SubjectId, e.TourId, e.TourVersionId, status)
		return wrap(err, domains.ErrTourNotFound, e.SubjectId, errs.Update)
	}
	return nil
}

func (rr *RuntimeRepository) Funnel(
	ctx context.Context,
	versionId uuid.UUID,
	from, to time.Time,
) ([]domains.FunnelStep, error) {
	rows, err := rr.pool.Query(ctx, `
		SELECT h.step, h.id, h.title,
		       COUNT(DISTINCT e.subject_id) FILTER (WHERE e.type = 'hint_shown')       AS shown,
		       COUNT(DISTINCT e.subject_id) FILTER (WHERE e.type = 'hint_completed')   AS completed,
		       COUNT(DISTINCT e.subject_id) FILTER (WHERE e.type = 'hint_skipped')     AS skipped,
		       COUNT(*)                     FILTER (WHERE e.type = 'selector_missing') AS selector_missing
		FROM hints h
		LEFT JOIN tour_events e
		       ON e.hint_id = h.id
		      AND e.tour_version_id = h.tour_version_id
		      AND e.received_at >= $2 AND e.received_at <= $3
		WHERE h.tour_version_id = $1
		GROUP BY h.step, h.id, h.title
		ORDER BY h.step`, versionId, from, to)
	if err != nil {
		return nil, wrap(err, domains.ErrVersionNotFound, versionId.String(), errs.Retrieve)
	}
	defer rows.Close()

	funnel := make([]domains.FunnelStep, 0)
	for rows.Next() {
		var s domains.FunnelStep
		if err := rows.Scan(&s.Step, &s.HintId, &s.Title,
			&s.Shown, &s.Completed, &s.Skipped, &s.SelectorMissing); err != nil {
			return nil, wrap(err, domains.ErrVersionNotFound, versionId.String(), errs.Retrieve)
		}
		funnel = append(funnel, s)
	}
	return funnel, rows.Err()
}

func (rr *RuntimeRepository) Totals(
	ctx context.Context,
	versionId uuid.UUID,
	from, to time.Time,
) (domains.AnalyticsTotals, error) {
	var t domains.AnalyticsTotals
	err := rr.pool.QueryRow(ctx, `
		SELECT
			COUNT(DISTINCT subject_id) FILTER (WHERE type = 'tour_started')   AS started,
			COUNT(DISTINCT subject_id) FILTER (WHERE type = 'tour_completed') AS completed,
			COUNT(DISTINCT subject_id) FILTER (WHERE type = 'tour_dismissed') AS dismissed,
			COUNT(DISTINCT subject_id) FILTER (WHERE type = 'goal_reached')   AS goal_reached
		FROM tour_events
		WHERE tour_version_id = $1 AND received_at >= $2 AND received_at <= $3`,
		versionId, from, to,
	).Scan(&t.Started, &t.Completed, &t.Dismissed, &t.GoalReached)

	return t, wrap(err, domains.ErrVersionNotFound, versionId.String(), errs.Retrieve)
}

func (rr *RuntimeRepository) GetSubjectEvents(ctx context.Context, appId uuid.UUID, subjectId string, since time.Time, limit int) ([]domains.Event, error) {
	if limit <= 0 {
		limit = 1000
	}
	query := `
		SELECT id, app_id, event_key, tour_id, tour_version_id, hint_id, session_id, subject_id, type, occurred_at, received_at, payload
		FROM tour_events
		WHERE app_id = $1 AND subject_id = $2`

	args := []any{appId, subjectId}
	argIdx := 3

	if !since.IsZero() {
		query += fmt.Sprintf(" AND occurred_at >= $%d", argIdx)
		args = append(args, since)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY occurred_at DESC LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := rr.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, wrap(err, domains.ErrTourNotFound, subjectId, errs.Retrieve)
	}
	defer rows.Close()

	events := make([]domains.Event, 0, limit)
	for rows.Next() {
		var e domains.Event
		var payloadBytes []byte
		var tourID, versionID *uuid.UUID

		if err := rows.Scan(
			&e.Id, &e.AppId, &e.EventKey, &tourID, &versionID, &e.HintId,
			&e.SessionId, &e.SubjectId, &e.Type, &e.OccurredAt, &e.ReceivedAt, &payloadBytes,
		); err != nil {
			return nil, wrap(err, domains.ErrTourNotFound, subjectId, errs.Retrieve)
		}

		if tourID != nil {
			e.TourId = *tourID
		}
		if versionID != nil {
			e.TourVersionId = *versionID
		}

		if len(payloadBytes) > 0 {
			var p map[string]any
			if err := json.Unmarshal(payloadBytes, &p); err != nil {
				rr.logger.Error().
					Int64("event_id", e.Id).
					Str("event_key", e.EventKey).
					Str("subject_id", e.SubjectId).
					Err(err).
					Msg("corrupted event payload in tour_events")
				e.Payload = make(map[string]any)
			} else {
				e.Payload = p
			}
		} else {
			e.Payload = make(map[string]any)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
