CREATE TABLE IF NOT EXISTS mibig.entries (
    entry_id text PRIMARY KEY,
    minimal bool NOT NULL,
    tax_id int NOT NULL,
    organism_name text NOT NULL,
    biosyn_class text[] NOT NULL,
    legacy_comment text
);

CREATE TABLE IF NOT EXISTS mibig.rel_entries_types (
    entry_id text REFERENCES mibig.entries ON DELETE CASCADE,
    bgc_type_id int REFERENCES mibig.bgc_types ON DELETE CASCADE,
    PRIMARY KEY (entry_id, bgc_type_id)
);
