CREATE SCHEMA IF NOT EXISTS live;

DO $$ BEGIN
    CREATE TYPE live.entry_status AS ENUM ('reserved', 'pending', 'active', 'retired');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE live.entry_quality AS ENUM ('questionable', 'low', 'medium', 'high');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE live.entry_completeness AS ENUM ('unknown', 'partial', 'complete');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS live.entries (
    entry_id text PRIMARY KEY,
    accession text NOT NULL,
    version int NOT NULL,
    status live.entry_status NOT NULL,
    quality live.entry_quality NOT NULL,
    completeness live.entry_completeness NOT NULL,
    tax_id bigint NOT NULL,
    organism_name text NOT NULL,
    retirement_reason text[],
    see_also text[],
    data jsonb NOT NULL
);

CREATE TABLE IF NOT EXISTS live.rel_entries_types (
    entry_id text REFERENCES live.entries ON DELETE CASCADE,
    bgc_type_id int REFERENCES data.bgc_types ON DELETE CASCADE,
    PRIMARY KEY (entry_id, bgc_type_id)
);

DO $$ BEGIN
    CREATE MATERIALIZED VIEW live.entry_bgc_info AS SELECT entry_id, array_agg(name) AS names, array_agg(description) AS descriptions, array_agg(safe_class) AS css_classes FROM live.entries, jsonb_to_recordset(live.entries.data -> 'biosynthesis' -> 'classes') AS specs(class text) LEFT JOIN data.bgc_types ON LOWER(class) = term GROUP BY entry_id ORDER BY entry_id;
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE MATERIALIZED VIEW live.entry_compounds AS SELECT entry_id, array_agg(name) AS compounds, array_agg(synonyms) filter(WHERE synonyms <> '{}') AS synonyms FROM live.entries, jsonb_to_recordset(live.entries.data -> 'compounds') AS specs(name text, synonyms jsonb)GROUP BY entry_id;
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;
