CREATE TYPE mibig.publication_type AS ENUM ('pubmed', 'doi', 'patent', 'url');

CREATE TABLE IF NOT EXISTS mibig.publications (
    pub_id serial PRIMARY KEY,
    pub_type mibig.publication_type NOT NULL,
    pub_reference text NOT NULL UNIQUE,
    abstract text
);

CREATE TABLE IF NOT EXISTS mibig.rel_entries_pubs (
    entry_id text NOT NULL REFERENCES mibig.entries ON DELETE CASCADE,
    pub_id int NOT NULL REFERENCES mibig.publications ON DELETE CASCADE,
    PRIMARY KEY (entry_id, pub_id)
);
