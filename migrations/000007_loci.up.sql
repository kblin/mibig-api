CREATE TYPE mibig.locus_evidence AS ENUM (
    'Sequence-based prediction',
    'Gene expression correlated with compound production',
    'Knock-out studies',
    'Enzymatic assays',
    'Heterologous expression'
);

CREATE TYPE mibig.locus_completeness AS ENUM (
    'incomplete',
    'unknown',
    'complete'
);

CREATE TABLE IF NOT EXISTS mibig.loci (
    locus_id serial PRIMARY KEY,
    accession text NOT NULL,
    completeness mibig.locus_completeness NOT NULL,
    evidences mibig.locus_evidence[],
    start_coord int,
    end_coord int,
    CHECK(start_coord < end_coord OR (start_coord = 0 AND end_coord = 0)),
    mixs_compliant bool,
    entry_id text REFERENCES mibig.entries ON DELETE CASCADE
);
