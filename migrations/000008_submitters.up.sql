CREATE SCHEMA IF NOT EXISTS mibig_submitters;

CREATE TABLE IF NOT EXISTS mibig_submitters.submitters (
    user_id text PRIMARY KEY,
    email citext UNIQUE NOT NULL,
    name text,
    call_name text,
    institution text,
    password_hash bytea,
    is_public boolean NOT NULL,
    gdpr_consent boolean NOT NULL,
    active boolean NOT NULL,
    version integer NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS mibig_submitters.roles (
    role_id int NOT NULL PRIMARY KEY,
    name TEXT UNIQUE,
    description TEXT
);

INSERT INTO mibig_submitters.roles (role_id, name, description)
VALUES
    (1, 'admin', 'Users who can manage other users'),
    (2, 'curator', 'Users who can approve new entries'),
    (3, 'submitter', 'Users who can edit entries'),
    (4, 'guest', 'Users with read only access');

CREATE TABLE IF NOT EXISTS mibig_submitters.rel_submitters_roles (
    user_id text REFERENCES mibig_submitters.submitters ON DELETE CASCADE,
    role_id int REFERENCES mibig_submitters.roles ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);
