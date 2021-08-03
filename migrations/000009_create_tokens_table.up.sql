CREATE TABLE IF NOT EXISTS mibig_submitters.tokens (
    hash bytea PRIMARY KEY,
    user_id text NOT NULL REFERENCES mibig_submitters.submitters ON DELETE CASCADE,
    expiry timestamp(0) WITH TIME ZONE NOT NULL,
    scope text NOT NULL
);