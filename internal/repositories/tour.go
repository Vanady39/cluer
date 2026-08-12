package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/Vanady39/cluer/internal/domains"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type TourRepository struct {
	pool   *pgxpool.Pool
	logger *zerolog.Logger
}

func NewTourRepository(pool *pgxpool.Pool, logger *zerolog.Logger) *TourRepository {
	return &TourRepository{pool: pool, logger: logger}
}

const versionColumns = `id, tour_id, version, status, trigger_type, target_path,
	audience_show_once, audience_max_shows, audience_only_new, 
	audience_rules, trigger_config,
	created_by, created_at, published_at, archived_at`

func (tr *TourRepository) CreateWithDraft(ctx context.Context, t *domains.Tour) (*domains.Tour, error) {
	tx, err := tr.pool.Begin(ctx)
	if err != nil {
		return nil, wrap(err, domains.ErrTourNotFound, "", domains.Create)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO tours (app_id, title, description, enabled, priority)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`,
		t.AppId, t.Title, t.Description, t.Enabled, t.Priority,
	).Scan(&t.Id, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, wrap(err, domains.ErrTourNotFound, t.Title, domains.Create)
	}

	triggerConfigBytes, err := marshalJSON(t.TriggerConfig)
	if err != nil {
		return nil, wrap(err, domains.ErrVersionNotFound, t.Id.String(), domains.Create)
	}

	var audienceRulesBytes []byte
	if t.Audience != nil && len(t.Audience.Rules) > 0 {
		audienceRulesBytes, err = marshalJSON(t.Audience.Rules)
		if err != nil {
			return nil, wrap(err, domains.ErrVersionNotFound, t.Id.String(), domains.Create)
		}
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO tour_versions (
			tour_id, version, status, trigger_type, target_path,
			audience_show_once, audience_max_shows, audience_only_new,
			trigger_config, audience_rules)
		VALUES ($1, 1, 'draft', $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, version`,
		t.Id, t.TriggerType, t.TargetPath,
		t.Audience.ShowOnce, t.Audience.MaxShows, t.Audience.OnlyNew,
		triggerConfigBytes, audienceRulesBytes,
	).Scan(&t.VersionId, &t.Version)
	if err != nil {
		return nil, wrap(err, domains.ErrVersionNotFound, t.Id.String(), domains.Create)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, wrap(err, domains.ErrTourNotFound, t.Id.String(), domains.Create)
	}

	t.Status = domains.TourDraft
	t.Hints = []domains.Hint{}
	return t, nil
}

func (tr *TourRepository) GetById(ctx context.Context, id uuid.UUID) (*domains.Tour, error) {
	t := new(domains.Tour)
	err := tr.pool.QueryRow(ctx, `
		SELECT id, app_id, title, description, enabled, priority, created_at, updated_at
		FROM tours
		WHERE id = $1 AND archived_at IS NULL`, id,
	).Scan(&t.Id, &t.AppId, &t.Title, &t.Description, &t.Enabled, &t.Priority, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, wrap(err, domains.ErrTourNotFound, id.String(), domains.Retrieve)
	}
	return t, nil
}

func (tr *TourRepository) List(ctx context.Context, appId uuid.UUID) ([]*domains.Tour, error) {
	rows, err := tr.pool.Query(ctx, `
		SELECT t.id, t.app_id, t.title, t.description, t.enabled, t.priority,
		       t.created_at, t.updated_at,
		       v.id, v.version, v.status, v.trigger_type, v.target_path,
		       v.audience_show_once, v.audience_max_shows, v.audience_only_new
		FROM tours t
		LEFT JOIN LATERAL (
			SELECT * FROM tour_versions tv
			WHERE tv.tour_id = t.id
			ORDER BY (tv.status = 'published') DESC, tv.version DESC
			LIMIT 1
		) v ON true
		WHERE t.app_id = $1 AND t.archived_at IS NULL
		ORDER BY t.priority DESC, t.created_at ASC`, appId)
	if err != nil {
		return nil, wrap(err, domains.ErrTourNotFound, appId.String(), domains.Retrieve)
	}
	defer rows.Close()

	tours := make([]*domains.Tour, 0)
	for rows.Next() {
		t := new(domains.Tour)
		var (
			versionId   *uuid.UUID
			version     *int
			status      *string
			triggerType *string
			targetPath  *string
			showOnce    *bool
			maxShows    *int
			onlyNew     *bool
		)
		if err := rows.Scan(
			&t.Id, &t.AppId, &t.Title, &t.Description, &t.Enabled, &t.Priority,
			&t.CreatedAt, &t.UpdatedAt,
			&versionId, &version, &status, &triggerType, &targetPath,
			&showOnce, &maxShows, &onlyNew,
		); err != nil {
			return nil, wrap(err, domains.ErrTourNotFound, appId.String(), domains.Retrieve)
		}

		if versionId != nil {
			t.VersionId = *versionId
			t.Version = *version
			t.Status = domains.TourStatus(*status)
			t.TriggerType = domains.TriggerType(*triggerType)
			t.TargetPath = *targetPath
			t.Audience = &domains.Audience{ShowOnce: *showOnce, MaxShows: *maxShows, OnlyNew: *onlyNew}
		}
		t.Hints = []domains.Hint{}
		tours = append(tours, t)
	}
	return tours, rows.Err()
}

func (tr *TourRepository) UpdateMeta(ctx context.Context, id uuid.UUID, p domains.TourMetaPatch) (*domains.Tour, error) {
	set := make([]string, 0, 4)
	args := make([]any, 0, 5)

	add := func(column string, value any) {
		args = append(args, value)
		set = append(set, fmt.Sprintf("%s = $%d", column, len(args)))
	}

	if p.Title != nil {
		add("title", *p.Title)
	}
	if p.Description != nil {
		add("description", *p.Description)
	}
	if p.Enabled != nil {
		add("enabled", *p.Enabled)
	}
	if p.Priority != nil {
		add("priority", *p.Priority)
	}
	if len(set) == 0 {
		return tr.GetById(ctx, id)
	}

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE tours SET %s
		WHERE id = $%d AND archived_at IS NULL
		RETURNING id, app_id, title, description, enabled, priority, created_at, updated_at`,
		strings.Join(set, ", "), len(args))

	t := new(domains.Tour)
	err := tr.pool.QueryRow(ctx, query, args...).Scan(
		&t.Id, &t.AppId, &t.Title, &t.Description, &t.Enabled, &t.Priority, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, wrap(err, domains.ErrTourNotFound, id.String(), domains.Update)
	}
	return t, nil
}

func (tr *TourRepository) Archive(ctx context.Context, id uuid.UUID) error {
	tag, err := tr.pool.Exec(ctx,
		`UPDATE tours SET archived_at = now() WHERE id = $1 AND archived_at IS NULL`, id)
	if err != nil {
		return wrap(err, domains.ErrTourNotFound, id.String(), domains.Delete)
	}
	if tag.RowsAffected() == 0 {
		return domains.NotFound(domains.ErrTourNotFound, id.String())
	}
	return nil
}

func (tr *TourRepository) GetVersion(ctx context.Context, versionId uuid.UUID) (*domains.TourVersion, error) {
	v, err := scanVersion(tr.pool.QueryRow(ctx,
		`SELECT `+versionColumns+` FROM tour_versions WHERE id = $1`, versionId))
	if err != nil {
		return nil, wrap(err, domains.ErrVersionNotFound, versionId.String(), domains.Retrieve)
	}
	return v, nil
}

func (tr *TourRepository) VersionOwner(ctx context.Context, versionId uuid.UUID) (uuid.UUID, uuid.UUID, error) {
	var tourId, appId uuid.UUID
	err := tr.pool.QueryRow(ctx, `
		SELECT v.tour_id, t.app_id
		FROM tour_versions v
		JOIN tours t ON t.id = v.tour_id
		WHERE v.id = $1`, versionId,
	).Scan(&tourId, &appId)
	if err != nil {
		return uuid.Nil, uuid.Nil, wrap(err, domains.ErrVersionNotFound, versionId.String(), domains.Retrieve)
	}
	return tourId, appId, nil
}

func (tr *TourRepository) ListVersions(ctx context.Context, tourId uuid.UUID) ([]*domains.TourVersion, error) {
	rows, err := tr.pool.Query(ctx,
		`SELECT `+versionColumns+` FROM tour_versions WHERE tour_id = $1 ORDER BY version DESC`, tourId)
	if err != nil {
		return nil, wrap(err, domains.ErrVersionNotFound, tourId.String(), domains.Retrieve)
	}
	defer rows.Close()

	versions := make([]*domains.TourVersion, 0)
	for rows.Next() {
		v, err := scanVersion(rows)
		if err != nil {
			return nil, wrap(err, domains.ErrVersionNotFound, tourId.String(), domains.Retrieve)
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func (tr *TourRepository) VersionByStatus(ctx context.Context, tourId uuid.UUID, s domains.TourStatus) (*domains.TourVersion, error) {
	v, err := scanVersion(tr.pool.QueryRow(ctx,
		`SELECT `+versionColumns+` FROM tour_versions WHERE tour_id = $1 AND status = $2`, tourId, s))
	if err != nil {
		return nil, wrap(err, domains.ErrVersionNotFound, tourId.String()+"/"+string(s), domains.Retrieve)
	}
	return v, nil
}

func (tr *TourRepository) UpdateDraft(ctx context.Context, versionId uuid.UUID, p domains.DraftPatch) (*domains.TourVersion, error) {
	set := make([]string, 0, 5)
	args := make([]any, 0, 6)

	add := func(column string, value any) {
		args = append(args, value)
		set = append(set, fmt.Sprintf("%s = $%d", column, len(args)))
	}

	if p.TriggerType != nil {
		add("trigger_type", *p.TriggerType)
	}
	if p.TargetPath != nil {
		add("target_path", *p.TargetPath)
	}
	if p.Audience != nil {
		add("audience_show_once", p.Audience.ShowOnce)
		add("audience_max_shows", p.Audience.MaxShows)
		add("audience_only_new", p.Audience.OnlyNew)

		rulesBytes, err := marshalJSON(p.Audience.Rules)

		if err != nil {
			return nil, wrap(err, domains.ErrVersionNotFound, versionId.String(), domains.Update)
		}

		add("audience_rules", rulesBytes)
	}

	if p.TriggerConfig != nil {
		configBytes, err := marshalJSON(p.TriggerConfig)
		if err != nil {
			return nil, wrap(err, domains.ErrVersionNotFound, versionId.String(), domains.Update)
		}
		add("trigger_config", configBytes)
	}

	if len(set) == 0 {
		return tr.GetVersion(ctx, versionId)
	}

	args = append(args, versionId)
	query := fmt.Sprintf(
		`UPDATE tour_versions SET %s WHERE id = $%d AND status = 'draft' RETURNING %s`,
		strings.Join(set, ", "), len(args), versionColumns)

	v, err := scanVersion(tr.pool.QueryRow(ctx, query, args...))
	if err != nil {
		return nil, wrap(err, domains.ErrVersionNotFound, versionId.String(), domains.Update)
	}
	return v, nil
}

func (tr *TourRepository) CopyToDraft(ctx context.Context, tourId, sourceVersionId uuid.UUID) (*domains.TourVersion, error) {
	tx, err := tr.pool.Begin(ctx)
	if err != nil {
		return nil, wrap(err, domains.ErrVersionNotFound, tourId.String(), domains.Create)
	}
	defer tx.Rollback(ctx)

	v, err := scanVersion(tx.QueryRow(ctx, `
		INSERT INTO tour_versions (
			tour_id, version, status, trigger_type, target_path,
			audience_show_once, audience_max_shows, audience_only_new,
		    trigger_config, audience_rules)
		SELECT src.tour_id,
		       (SELECT COALESCE(MAX(version), 0) + 1 FROM tour_versions WHERE tour_id = $1),
		       'draft', src.trigger_type, src.target_path,
		       src.audience_show_once, src.audience_max_shows, src.audience_only_new,
		       src.trigger_config, src.audience_rules
		FROM tour_versions src
		WHERE src.id = $2 AND src.tour_id = $1
		RETURNING `+versionColumns, tourId, sourceVersionId))
	if err != nil {
		return nil, wrap(err, domains.ErrVersionNotFound, sourceVersionId.String(), domains.Create)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO hints (
			tour_version_id, step, title, content, selector, placement, page_path,
			media_url, spotlight, required, wait_for_selector, input_placeholder, expected_input)
		SELECT $1, step, title, content, selector, placement, page_path,
		       media_url, spotlight, required, wait_for_selector, input_placeholder, expected_input
		FROM hints WHERE tour_version_id = $2
		ORDER BY step`, v.Id, sourceVersionId)
	if err != nil {
		return nil, wrap(err, domains.ErrHintNotFound, v.Id.String(), domains.Create)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, wrap(err, domains.ErrVersionNotFound, v.Id.String(), domains.Create)
	}
	return v, nil
}

func (tr *TourRepository) Publish(ctx context.Context, tourId uuid.UUID) (*domains.TourVersion, error) {
	tx, err := tr.pool.Begin(ctx)
	if err != nil {
		return nil, wrap(err, domains.ErrTourNotFound, tourId.String(), domains.Update)
	}
	defer tx.Rollback(ctx)

	var locked uuid.UUID
	if err := tx.QueryRow(ctx,
		`SELECT id FROM tours WHERE id = $1 AND archived_at IS NULL FOR UPDATE`, tourId,
	).Scan(&locked); err != nil {
		return nil, wrap(err, domains.ErrTourNotFound, tourId.String(), domains.Update)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE tour_versions SET status = 'archived', archived_at = now()
		WHERE tour_id = $1 AND status = 'published'`, tourId); err != nil {
		return nil, wrap(err, domains.ErrVersionNotFound, tourId.String(), domains.Update)
	}

	v, err := scanVersion(tx.QueryRow(ctx, `
		UPDATE tour_versions SET status = 'published', published_at = now()
		WHERE tour_id = $1 AND status = 'draft'
		RETURNING `+versionColumns, tourId))
	if err != nil {
		return nil, wrap(err, domains.ErrNoDraft, tourId.String(), domains.Update)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, wrap(err, domains.ErrVersionNotFound, tourId.String(), domains.Update)
	}
	return v, nil
}

func marshalJSON(v any) ([]byte, error) {
	if v == nil {
		return nil, nil
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		if rv.IsNil() {
			return nil, nil
		}
	}

	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("json marshal: %w", err)
	}
	return b, nil
}

func unmarshalJSONB(data []byte, dst any) error {
	if len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("%w: %v", domains.ErrCorruptedData, err)
	}
	return nil
}

func (tr *TourRepository) ResolveCandidates(ctx context.Context, appId uuid.UUID) ([]*domains.Tour, error) {
	rows, err := tr.pool.Query(ctx, `
		SELECT t.id, t.app_id, t.title, t.description, t.enabled, t.priority,
		       t.created_at, t.updated_at,
		       v.id, v.version, v.trigger_type, v.target_path,
		       v.audience_show_once, v.audience_max_shows, v.audience_only_new,
		       v.audience_rules, v.trigger_config
		FROM tours t
		JOIN tour_versions v ON v.tour_id = t.id AND v.status = 'published'
		WHERE t.app_id = $1 AND t.enabled AND t.archived_at IS NULL
		ORDER BY t.priority DESC, t.created_at ASC`, appId)
	if err != nil {
		return nil, wrap(err, domains.ErrTourNotFound, appId.String(), domains.Retrieve)
	}
	defer rows.Close()

	tours := make([]*domains.Tour, 0)
	for rows.Next() {
		t := new(domains.Tour)
		t.Audience = new(domains.Audience)

		var audienceRulesBytes []byte
		var triggerConfigBytes []byte

		if err := rows.Scan(
			&t.Id, &t.AppId, &t.Title, &t.Description, &t.Enabled, &t.Priority,
			&t.CreatedAt, &t.UpdatedAt,
			&t.VersionId, &t.Version, &t.TriggerType, &t.TargetPath,
			&t.Audience.ShowOnce, &t.Audience.MaxShows, &t.Audience.OnlyNew,
			&audienceRulesBytes, &triggerConfigBytes,
		); err != nil {
			tr.logger.Error().
				Err(err).
				Str("app_id", appId.String()).
				Msg("failed to scan tour row in ResolveCandidates, skipping")
			continue
		}
		if err := unmarshalJSONB(audienceRulesBytes, &t.Audience.Rules); err != nil {
			tr.logger.Error().
				Err(err).
				Str("tour_id", t.Id.String()).
				Str("version_id", t.VersionId.String()).
				Str("app_id", appId.String()).
				Msg("corrupted audience_rules in published tour, skipping tour")
			continue
		}

		if err := unmarshalJSONB(triggerConfigBytes, &t.TriggerConfig); err != nil {
			tr.logger.Error().
				Err(err).
				Str("tour_id", t.Id.String()).
				Str("version_id", t.VersionId.String()).
				Str("app_id", appId.String()).
				Msg("corrupted trigger_config in published tour, skipping tour")
			continue
		}

		t.Status = domains.TourPublished
		t.Hints = []domains.Hint{}
		tours = append(tours, t)
	}
	return tours, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanVersion(row scanner) (*domains.TourVersion, error) {
	v := new(domains.TourVersion)

	var createdBy *string

	var triggerConfigBytes []byte
	var audienceRulesBytes []byte
	err := row.Scan(
		&v.Id, &v.TourId, &v.Version, &v.Status, &v.TriggerType, &v.TargetPath,
		&v.Audience.ShowOnce, &v.Audience.MaxShows, &v.Audience.OnlyNew,
		&audienceRulesBytes,
		&triggerConfigBytes,
		&createdBy, &v.CreatedAt, &v.PublishedAt, &v.ArchivedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := unmarshalJSONB(triggerConfigBytes, &v.TriggerConfig); err != nil {
		return nil, fmt.Errorf("scanVersion trigger_config (version %s): %w", v.Id, err)
	}
	if err := unmarshalJSONB(audienceRulesBytes, &v.Audience.Rules); err != nil {
		return nil, fmt.Errorf("scanVersion audience_rules (version %s): %w", v.Id, err)
	}

	if createdBy != nil {
		v.CreatedBy = *createdBy
	}
	return v, nil
}
