DROP TABLE IF EXISTS listings;
DROP TABLE IF EXISTS demo_seed_state;

DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS tour_events;
DROP TABLE IF EXISTS tour_progress;

DROP TRIGGER IF EXISTS hints_immutability ON hints;
DROP TRIGGER IF EXISTS hints_set_updated_at ON hints;
DROP TABLE IF EXISTS hints;
DROP FUNCTION IF EXISTS hints_freeze();

DROP TRIGGER IF EXISTS tour_versions_immutability ON tour_versions;
DROP TABLE IF EXISTS tour_versions;
DROP FUNCTION IF EXISTS tour_versions_freeze();

DROP TRIGGER IF EXISTS tours_set_updated_at ON tours;
DROP TABLE IF EXISTS tours;

DROP TABLE IF EXISTS apps;

DROP FUNCTION IF EXISTS set_updated_at();
