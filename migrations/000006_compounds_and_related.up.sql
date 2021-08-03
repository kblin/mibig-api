CREATE TABLE IF NOT EXISTS mibig.chem_compounds (
    compound_id serial PRIMARY KEY,
    entry_id text NOT NULL REFERENCES mibig.entries ON DELETE CASCADE,
    activities text[],
    structure text,
    synonyms text[],
    name text,
    database_ids text[],
    -- TODO: Turn into an ENUM
    evidences text[],
    mass_spec_ion_type text,
    molecuar_mass double precision,
    molecular_formula text
);

CREATE TABLE IF NOT EXISTS mibig.chem_targets (
    chem_target_id serial PRIMARY KEY,
    name text NOT NULL,
    compound_id int REFERENCES mibig.chem_compounds ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS mibig.chem_moieties (
    chem_moiety_id serial PRIMARY KEY,
    -- TODO: Add references to subclusters
    name text NOT NULL,
    compound_id int REFERENCES mibig.chem_compounds ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS mibig.rel_targets_pubs (
    chem_target_id int NOT NULL REFERENCES mibig.chem_targets ON DELETE CASCADE,
    pub_id int NOT NULL REFERENCES mibig.publications ON DELETE CASCADE,
    PRIMARY KEY (chem_target_id, pub_id)
);
