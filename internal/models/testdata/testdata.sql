INSERT INTO mibig.entries (entry_id, minimal, ncbi_taxid, organism_name, biosyn_class ) VALUES
('BGC0000535', FALSE, 1360, 'Lactococcus lactis subsp. lactis', '{RiPP}'),
('BGC0001070', FALSE, 1214242, 'Streptomyces collinus Tu 365', '{NRP,Polyketide}');

INSERT INTO mibig.changelogs (comments, version_id, entry_id) VALUES
('{Submitted}', 1, 'BGC0000535'),
('{"Migrated from v1.4"}', 6, 'BGC0000535'),
('{Submitted}', 1, 'BGC0001070'),
('{"Migrated from v1.4"}', 6, 'BGC0001070');

INSERT INTO mibig.chem_compounds (entry_id, synonyms, name) VALUES
('BGC0000535', '{nisin}', 'nisin A'),
('BGC0001070', '{mocimycin,delvomycin}', 'kirromycin');

INSERT INTO mibig.loci (accession, completeness, evidences, start_coord, end_coord, mixs_compliant, entry_id) VALUES
('HM219853.1', 'complete', '{"Gene expression correlated with compound production","Sequence-based prediction"}', 0, 0, FALSE, 'BGC0000535'),
('AM746336.1', 'complete', '{"Knock-out studies"}', 0, 0, FALSE, 'BGC0001070');

INSERT INTO mibig.publications (pub_type, pub_reference) VALUES
('pubmed', '21183019'),
('pubmed', '4554808'),
('pubmed', '4373734'),
('pubmed', '18291322'),
('pubmed', '21513880'),
('pubmed', '19609288'),
('pubmed', '21401713'),
('pubmed', '23828654');

INSERT INTO mibig.rel_entries_pubs (entry_id, pub_id) VALUES
('BGC0000535', 1),
('BGC0001070', 2),
('BGC0001070', 3),
('BGC0001070', 4),
('BGC0001070', 5),
('BGC0001070', 6),
('BGC0001070', 7),
('BGC0001070', 8);

INSERT INTO mibig.rel_entries_types (entry_id, bgc_type_id) VALUES
('BGC0000535', 32),
('BGC0001070', 2),
('BGC0001070', 15),
('BGC0001070', 13);

INSERT INTO mibig.taxa (ncbi_taxid, superkingdom, phylum, class, taxonomic_order, family, genus, species, name) VALUES
(1360, 'Bacteria', 'Firmicutes', 'Bacilli', 'Lactobacillales', 'Streptococcaceae', 'Lactococcus', 'lactis', 'Lactococcus lactis subsp. lactis'),
(1214242, 'Bacteria', 'Actinobacteria', 'Actinobacteria', 'Streptomycetales', 'Streptomycetaceae', 'Streptomyces', 'collinus', 'Streptomyces collinus Tu 365');

INSERT INTO mibig_submitters.submitters (
    user_id, email, name, call_name, institution, password_hash, is_public, gdpr_consent, active
) VALUES
('AAAAAAAAAAAAAAAAAAAAAAAA', 'mibig@example.org', 'MIBiG Submitters', 'MIBiG', 'MIBiG', 'unused', TRUE, TRUE, FALSE),
('AAAAAAAAAAAAAAAAAAAAAAAB', 'alice@example.org', 'Alice User', 'Alice', 'Testing', 'unused', TRUE, TRUE, TRUE),
('AAAAAAAAAAAAAAAAAAAAAAAC', 'bob@example.org', 'Bob User', 'Bob', 'Testing', 'unused', TRUE, FALSE, TRUE),
('AAAAAAAAAAAAAAAAAAAAAAAD', 'chuck@example.org', 'Chuck User', 'Chuck', 'Testing', 'unused', FALSE, FALSE, TRUE);

INSERT INTO mibig_submitters.rel_submitters_roles (user_id, role_id) VALUES
('AAAAAAAAAAAAAAAAAAAAAAAA', 4),
('AAAAAAAAAAAAAAAAAAAAAAAB', 1),
('AAAAAAAAAAAAAAAAAAAAAAAC', 2),
('AAAAAAAAAAAAAAAAAAAAAAAD', 3);
