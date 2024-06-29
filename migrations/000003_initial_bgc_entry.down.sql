DROP MATERIALIZED VIEW IF EXISTS live.entry_compounds;
DROP MATERIALIZED VIEW IF EXISTS live.entry_bgc_info;

DROP TABLE IF EXISTS live.rel_entries_types;

DROP TABLE IF EXISTS live.entries;

DROP TYPE IF EXISTS live.entry_completeness;
DROP TYPE IF EXISTS live.entry_quality;
DROP TYPE IF EXISTS live.entry_status;

DROP SCHEMA IF EXISTS live;
