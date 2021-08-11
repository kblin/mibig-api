CREATE SCHEMA IF NOT EXISTS mibig_submissions;

CREATE TABLE IF NOT EXISTS mibig_submissions.next_accession (
    accession int PRIMARY KEY,
    version int
);


CREATE TABLE IF NOT EXISTS mibig_submissions.accession_requests (
    id serial PRIMARY KEY,
    user_id text NOT NULL REFERENCES mibig_submitters.submitters ON DELETE CASCADE,
    compounds text[] NOT NULL
);

CREATE TABLE IF NOT EXISTS mibig_submissions.accession_request_loci (
    id serial PRIMARY KEY,
    accession text NOT NULL,
    start_pos int NOT NULL,
    end_pos int NOT NULL,
    request int NOT NULL REFERENCES mibig_submissions.accession_requests ON DELETE CASCADE
);