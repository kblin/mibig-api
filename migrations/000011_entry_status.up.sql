CREATE TYPE mibig.entry_status AS ENUM ('published', 'retired', 'embargoed', 'reserved');

CREATE TABLE IF NOT EXISTS mibig.entry_status (
    entry_id text PRIMARY KEY,
    status mibig.entry_status NOT NULL,
    reason text NOT NULL,
    see text[]
);