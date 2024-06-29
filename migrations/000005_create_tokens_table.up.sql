CREATE TABLE IF NOT EXISTS auth.tokens (
    hash bytea PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES auth.users ON DELETE CASCADE,
    expiry timestamp(0) WITH TIME ZONE NOT NULL,
    scope text NOT NULL
);
