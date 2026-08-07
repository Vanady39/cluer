package repositories

import (
	"context"
	"time"

	"github.com/Vanady39/cluer/internal/domains"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RuntimeRepository struct {
	pool *pgxpool.Pool
}

func NewRuntimeRepository(pool *pgxpool.Pool) *RuntimeRepository {
	return &RuntimeRepository{pool: pool}
}

func (rr *RuntimeRepository) Ping(ctx context.Context) error {
	return rr.pool.Ping(ctx)
}

// ---------------------------------------------------------------- apps

func (rr *RuntimeRepository) CreateApp(ctx context.Context, a *domains.App) error {
	err := rr.pool.QueryRow(ctx, `
		INSERT INTO apps (name, public_key, allowed_origins)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`,
		a.Name, a.PublicKey, a.AllowedOrigins,
	).Scan(&a.Id, &a.CreatedAt)

	return wrap(err, domains.ErrAppNotFound, a.PublicKey, domains.Create)
}

func (rr *RuntimeRepository) AppByKey(ctx context.Context, key string) (*domains.App, error) {
	a := new(domains.App)
	err := rr.pool.QueryRow(ctx, `
		SELECT id, name, public_key, allowed_origins, created_at, archived_at
		FROM apps
		WHERE public_key = $1 AND archived_at IS NULL`, key,
	).Scan(&a.Id, &a.Name, &a.PublicKey, &a.AllowedOrigins, &a.CreatedAt, &a.ArchivedAt)
	if err != nil {
		return nil, wrap(err, domains.ErrAppNotFound, key, domains.Retrieve)
	}
	return a, nil
}

func (rr *RuntimeRepository) ListApps(ctx context.Context) ([]*domains.App, error) {
	rows, err := rr.pool.Query(ctx, `
		SELECT id, name, public_key, allowed_origins, created_at, archived_at
		FROM apps WHERE archived_at IS NULL ORDER BY created_at`)
	if err != nil {
		return nil, wrap(err, domains.ErrAppNotFound, "", domains.Retrieve)
	}
	defer rows.Close()

	apps := make([]*domains.App, 0)
	for rows.Next() {
		a := new(domains.App)
		if err := rows.Scan(&a.Id, &a.Name, &a.PublicKey, &a.AllowedOrigins,
			&a.CreatedAt, &a.ArchivedAt); err != nil {
			return nil, wrap(err, domains.ErrAppNotFound, "", domains.Retrieve)
		}
		apps = append(apps, a)
	}
	return apps, rows.Err()
}

// ---------------------------------------------------------------- progress

// SubjectProgress loads everything this subject has going, keyed by tour. One
// query rather than one per candidate: resolve has to answer in milliseconds,
// and the primary key covers the (app_id, subject_id) prefix.
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
		return nil, wrap(err, domains.ErrTourNotFound, subjectId, domains.Retrieve)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]*domains.Progress)
	for rows.Next() {
		p := new(domains.Progress)
		if err := rows.Scan(&p.AppId, &p.SubjectId, &p.TourId, &p.TourVersionId,
			&p.CurrentHintId, &p.Status, &p.ShowsCount,
			&p.StartedAt, &p.UpdatedAt, &p.FinishedAt); err != nil {
			return nil, wrap(err, domains.ErrTourNotFound, subjectId, domains.Retrieve)
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

	return wrap(err, domains.ErrTourNotFound, p.SubjectId, domains.Update)
}

// ---------------------------------------------------------------- events

// InsertEvents writes the batch and applies its progress side effects in one
// transaction.
//
// ON CONFLICT DO NOTHING is what makes a retry safe: the SDK resends batches
// after a dropped connection, and without it every retry would double the
// metrics. A row that conflicts is counted as a duplicate and, importantly,
// does not move the progress a second time — idempotent inserts would be
// worthless if the side effects still ran twice.
func (rr *RuntimeRepository) InsertEvents(ctx context.Context, events []domains.Event) (int, int, error) {
	tx, err := rr.pool.Begin(ctx)
	if err != nil {
		return 0, 0, wrap(err, domains.ErrTourNotFound, "", domains.Create)
	}
	defer tx.Rollback(ctx)

	accepted, duplicates := 0, 0

	for i := range events {
		e := events[i]

		tag, err := tx.Exec(ctx, `
			INSERT INTO tour_events (
				app_id, event_key, tour_id, tour_version_id, hint_id,
				session_id, subject_id, type, occurred_at, payload)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			ON CONFLICT (app_id, event_key) DO NOTHING`,
			e.AppId, e.EventKey, e.TourId, e.TourVersionId, e.HintId,
			e.SessionId, e.SubjectId, e.Type, e.OccurredAt, e.Payload)
		if err != nil {
			return 0, 0, wrap(err, domains.ErrTourNotFound, e.EventKey, domains.Create)
		}

		if tag.RowsAffected() == 0 {
			duplicates++
			continue
		}
		accepted++

		if err := applyProgress(ctx, tx, &e); err != nil {
			return 0, 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, 0, wrap(err, domains.ErrTourNotFound, "", domains.Create)
	}
	return accepted, duplicates, nil
}

// applyProgress is the denormalised projection of the event log. goal_reached is
// deliberately absent: a business outcome is not a step of the tour, and folding
// it into completion would make the two metrics measure each other.
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
			  AND status = 'in_progress'`,
			e.AppId, e.SubjectId, e.TourId, e.TourVersionId, e.HintId)
		return wrap(err, domains.ErrTourNotFound, e.SubjectId, domains.Update)

	case domains.EventTourCompleted, domains.EventTourDismissed:
		status := domains.ProgressCompleted
		if e.Type == domains.EventTourDismissed {
			status = domains.ProgressDismissed
		}
		_, err := tx.Exec(ctx, `
			UPDATE tour_progress
			SET status = $4, finished_at = now(), current_hint_id = NULL
			WHERE app_id = $1 AND subject_id = $2 AND tour_id = $3
			  AND status = 'in_progress'`,
			e.AppId, e.SubjectId, e.TourId, status)
		return wrap(err, domains.ErrTourNotFound, e.SubjectId, domains.Update)
	}
	return nil
}

// ---------------------------------------------------------------- analytics

// Funnel counts one row per hint of the version.
//
// The join is a LEFT JOIN so a hint nobody ever reached still shows up with
// zeros — that is precisely the row the admin needs to see. The order comes from
// hints.step rather than from the data, so the funnel reads top to bottom even
// when the middle of it is empty.
//
// COUNT(DISTINCT subject_id) rather than COUNT(*): a user who comes back to a
// hint twice is one person, and the funnel measures people.
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
		      AND e.received_at >= $2 AND e.received_at <= $3
		WHERE h.tour_version_id = $1
		GROUP BY h.step, h.id, h.title
		ORDER BY h.step`, versionId, from, to)
	if err != nil {
		return nil, wrap(err, domains.ErrVersionNotFound, versionId.String(), domains.Retrieve)
	}
	defer rows.Close()

	funnel := make([]domains.FunnelStep, 0)
	for rows.Next() {
		var s domains.FunnelStep
		if err := rows.Scan(&s.Step, &s.HintId, &s.Title,
			&s.Shown, &s.Completed, &s.Skipped, &s.SelectorMissing); err != nil {
			return nil, wrap(err, domains.ErrVersionNotFound, versionId.String(), domains.Retrieve)
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

	return t, wrap(err, domains.ErrVersionNotFound, versionId.String(), domains.Retrieve)
}
