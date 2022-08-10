CREATE TYPE mibig.mibig_status AS ENUM ('reserved', 'pending', 'active', 'retired');

CREATE TABLE IF NOT EXISTS mibig.entries (
    entry_id text PRIMARY KEY,
    minimal bool NOT NULL,
    tax_id bigint NOT NULL,
    organism_name text NOT NULL,
    biosyn_class text[] NOT NULL,
    status mibig.mibig_status NOT NULL,
    retirement_reason text[],
    see_also text[],
    legacy_comment text
);

CREATE TABLE IF NOT EXISTS mibig.rel_entries_types (
    entry_id text REFERENCES mibig.entries ON DELETE CASCADE,
    bgc_type_id int REFERENCES mibig.bgc_types ON DELETE CASCADE,
    PRIMARY KEY (entry_id, bgc_type_id)
);
