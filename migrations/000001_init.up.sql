
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
  NEW.updated_at := now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;



CREATE TABLE apps (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  name            TEXT        NOT NULL,
  public_key      TEXT        NOT NULL UNIQUE,   
  allowed_origins TEXT[]      NOT NULL DEFAULT '{}',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  archived_at     TIMESTAMPTZ
);

CREATE INDEX apps_active_key_idx ON apps (public_key) WHERE archived_at IS NULL;



CREATE TABLE tours (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  app_id      UUID        NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
  title       TEXT        NOT NULL,
  description TEXT        NOT NULL DEFAULT '',
  enabled     BOOLEAN     NOT NULL DEFAULT false,
  priority    INT         NOT NULL DEFAULT 0,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  archived_at TIMESTAMPTZ                       
);

CREATE INDEX tours_resolve_idx
  ON tours (app_id, priority DESC, created_at ASC)
  WHERE enabled AND archived_at IS NULL;

CREATE TRIGGER tours_set_updated_at
  BEFORE UPDATE ON tours
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();



CREATE TABLE tour_versions (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  tour_id      UUID        NOT NULL REFERENCES tours(id) ON DELETE RESTRICT,
  version      INT         NOT NULL,
  status       TEXT        NOT NULL CHECK (status IN ('draft','published','archived')),

  trigger_type TEXT        NOT NULL DEFAULT 'on_load',
  target_path  TEXT        NOT NULL,
  audience_show_once BOOLEAN NOT NULL DEFAULT false,
  audience_max_shows INT     NOT NULL DEFAULT 0,  
  audience_only_new  BOOLEAN NOT NULL DEFAULT false,
  -- ---------------------------------------------------------------

  created_by   TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  published_at TIMESTAMPTZ,
  archived_at  TIMESTAMPTZ,
  UNIQUE (tour_id, version),
  CHECK (status <> 'published' OR published_at IS NOT NULL),
  CHECK (status <> 'archived'  OR archived_at  IS NOT NULL)
);

CREATE UNIQUE INDEX tour_versions_one_published
  ON tour_versions (tour_id) WHERE status = 'published';

CREATE UNIQUE INDEX tour_versions_one_draft
  ON tour_versions (tour_id) WHERE status = 'draft';

CREATE INDEX tour_versions_history_idx ON tour_versions (tour_id, version DESC);


CREATE TABLE hints (
  id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  tour_version_id   UUID        NOT NULL REFERENCES tour_versions(id) ON DELETE CASCADE,
  step              INT         NOT NULL,
  title             TEXT        NOT NULL,
  content           TEXT        NOT NULL,
  selector          TEXT        NOT NULL DEFAULT '',
  placement         TEXT        NOT NULL,
  page_path         TEXT        NOT NULL DEFAULT '',
  media_url         TEXT        NOT NULL DEFAULT '',
  spotlight         BOOLEAN     NOT NULL DEFAULT false,
  required          BOOLEAN     NOT NULL DEFAULT false,
  wait_for_selector BOOLEAN     NOT NULL DEFAULT false,
  input_placeholder TEXT        NOT NULL DEFAULT '',
  expected_input    TEXT        NOT NULL DEFAULT '',
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tour_version_id, step) DEFERRABLE INITIALLY IMMEDIATE
);

CREATE TRIGGER hints_set_updated_at
  BEFORE UPDATE ON hints
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();


CREATE OR REPLACE FUNCTION tour_versions_freeze() RETURNS trigger AS $$
BEGIN
  IF OLD.status <> 'draft'
     AND (NEW.tour_id            <> OLD.tour_id
       OR NEW.version            <> OLD.version
       OR NEW.trigger_type       <> OLD.trigger_type
       OR NEW.target_path        <> OLD.target_path
       OR NEW.audience_show_once <> OLD.audience_show_once
       OR NEW.audience_max_shows <> OLD.audience_max_shows
       OR NEW.audience_only_new  <> OLD.audience_only_new) THEN
    RAISE EXCEPTION 'tour_version %: body of a % version is immutable', OLD.id, OLD.status
      USING ERRCODE = 'restrict_violation';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tour_versions_immutability
  BEFORE UPDATE ON tour_versions
  FOR EACH ROW EXECUTE FUNCTION tour_versions_freeze();

CREATE OR REPLACE FUNCTION hints_freeze() RETURNS trigger AS $$
DECLARE
  parent UUID := COALESCE(NEW.tour_version_id, OLD.tour_version_id);
  st     TEXT;
BEGIN
  SELECT status INTO st FROM tour_versions WHERE id = parent;

  IF NOT FOUND THEN
    RETURN COALESCE(NEW, OLD);
  END IF;

  IF st <> 'draft' THEN
    RAISE EXCEPTION 'hints of a % version are immutable (version %)', st, parent
      USING ERRCODE = 'restrict_violation';
  END IF;

  RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER hints_immutability
  BEFORE INSERT OR UPDATE OR DELETE ON hints
  FOR EACH ROW EXECUTE FUNCTION hints_freeze();


CREATE TABLE tour_progress (
  app_id          UUID        NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
  subject_id      TEXT        NOT NULL,   -- userId хоста либо anonId из localStorage
  tour_id         UUID        NOT NULL REFERENCES tours(id) ON DELETE RESTRICT,
  tour_version_id UUID        NOT NULL REFERENCES tour_versions(id) ON DELETE RESTRICT,
  current_hint_id UUID        REFERENCES hints(id) ON DELETE RESTRICT,
  status          TEXT        NOT NULL,
  shows_count     INT         NOT NULL DEFAULT 0,
  started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  finished_at     TIMESTAMPTZ,
  PRIMARY KEY (app_id, subject_id, tour_id),
  CHECK ((finished_at IS NULL) = (status = 'in_progress'))
);

CREATE INDEX tour_progress_tour_status_idx ON tour_progress (tour_id, status);

CREATE TRIGGER tour_progress_set_updated_at
  BEFORE UPDATE ON tour_progress
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();


CREATE TABLE tour_events (
  id              BIGSERIAL   PRIMARY KEY,
  app_id          UUID        NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
  event_key       TEXT        NOT NULL,   
  tour_id         UUID        NOT NULL,  
  tour_version_id UUID        NOT NULL REFERENCES tour_versions(id) ON DELETE RESTRICT,
  hint_id         UUID        REFERENCES hints(id) ON DELETE RESTRICT,
  session_id      TEXT        NOT NULL,
  subject_id      TEXT        NOT NULL,
  type            TEXT        NOT NULL,
  occurred_at     TIMESTAMPTZ NOT NULL,   
  received_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  payload         JSONB       NOT NULL DEFAULT '{}',
  UNIQUE (app_id, event_key)              
);

CREATE INDEX tour_events_version_type_idx
  ON tour_events (tour_version_id, type, received_at);

CREATE INDEX tour_events_subject_idx
  ON tour_events (app_id, subject_id, occurred_at);


CREATE TABLE reports (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  app_id          UUID        NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
  tour_id         UUID        NOT NULL REFERENCES tours(id) ON DELETE RESTRICT,
  tour_version_id UUID        NOT NULL REFERENCES tour_versions(id) ON DELETE RESTRICT,
  period_from     TIMESTAMPTZ NOT NULL,
  period_to       TIMESTAMPTZ NOT NULL,
  status          TEXT        NOT NULL,
  file_path       TEXT,
  error           TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (period_to > period_from)
);

CREATE INDEX reports_tour_created_idx ON reports (tour_id, created_at DESC);


CREATE TABLE listings (
  id          BIGINT      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  title       TEXT        NOT NULL,
  description TEXT        NOT NULL DEFAULT '',
  price       BIGINT      NOT NULL CHECK (price >= 0),
  image_url   TEXT        NOT NULL DEFAULT '',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
