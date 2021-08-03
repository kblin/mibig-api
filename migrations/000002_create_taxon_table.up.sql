CREATE TABLE mibig.taxa (
    tax_id	serial PRIMARY KEY,
    ncbi_taxid	bigint NOT NULL,
    superkingdom	text,
    kingdom	text,
    phylum	text,
    class	text,
    taxonomic_order	text,
    family	text,
    genus	text,
    species	text,
    name	text UNIQUE NOT NULL
);
