CREATE EXTENSION IF NOT EXISTS citext;

CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE IF NOT EXISTS auth.users (
    user_id bigserial PRIMARY KEY,
    email citext UNIQUE NOT NULL,
    active boolean NOT NULL,
    password_hash text NOT NULL,
    version integer NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS auth.user_info (
    user_id bigint REFERENCES auth.users ON DELETE CASCADE,
    alias text NOT NULL,
    name text,
    call_name text,
    organisation_1 text,
    organisation_2 text,
    organisation_3 text,
    orcid text,
    public boolean NOT NULL,
    version integer NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS auth.roles (
    role_id bigserial PRIMARY KEY,
    name TEXT UNIQUE,
    description TEXT
);

INSERT INTO auth.roles (role_id, name, description)
VALUES
    (1, 'submitter', 'Users who can edit entries'),
    (2, 'reviewer', 'Users who can approve new entries'),
    (3, 'admin', 'Users who can manage other users');

CREATE TABLE IF NOT EXISTS auth.rel_user_roles (
    user_id bigint REFERENCES auth.users ON DELETE CASCADE,
    role_id bigint REFERENCES auth.roles ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);
