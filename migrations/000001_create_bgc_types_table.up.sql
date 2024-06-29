CREATE SCHEMA IF NOT EXISTS data;

CREATE TABLE data.bgc_types (
    bgc_type_id	serial NOT NULL,
    term	text,
    name	text,
    description	text,
    safe_class	text,
    CONSTRAINT bgc_types_pkey PRIMARY KEY (bgc_type_id),
    CONSTRAINT bgc_types_term_unique UNIQUE (term),
    CONSTRAINT bgc_types_name_unique UNIQUE (name)
);

COMMENT ON TABLE data.bgc_types IS
  'Biosynthetic gene cluster types.';

--- basic MIBiG types
INSERT INTO data.bgc_types (term, name, description, safe_class)
SELECT val.term, val.name, val.description, val.safe_class
FROM (
    VALUES
        ('nrps', 'NRP', 'Nonribosomal peptide', 'nrps'),
        ('pks', 'Polyketide', 'Polyketide', 'pks'),
        ('ribosomal', 'Ribosomal', 'Ribosomally synthesized peptide', 'ripp'),
        ('saccharide', 'Saccharide', 'Saccharide', 'saccharide'),
        ('terpene', 'Terpene', 'Terpene', 'terpene'),
        ('other', 'Other', 'Other', 'other')
    ) val ( term, name, description, safe_class );
