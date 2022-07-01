package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/lib/pq"
	"secondarymetabolites.org/mibig-api/internal/data"
)

func (m *LiveEntryModel) Add(entry data.MibigEntry) error {
	ctx := context.Background()
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = insertEntry(entry, ctx, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = insertEntryBgcRelation(entry, ctx, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = insertChangeLog(entry, ctx, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = insertPublications(entry, ctx, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = insertCompounds(entry, ctx, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = insertLocus(entry, ctx, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (m LiveEntryModel) InsertEntryStatus(status data.MibigEntryStatus) error {
	query := `INSERT INTO mibig.entry_status (entry_id, status, reason, see)
	VALUES ($1, $2, $3, $4)`

	args := []interface{}{
		status.EntryId,
		status.Status,
		status.Reason,
		pq.Array(status.See),
	}

	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_DB_TIMEOUT)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func insertEntry(entry data.MibigEntry, ctx context.Context, tx *sql.Tx) error {

	ncbi_taxid, err := strconv.ParseInt(entry.Cluster.NcbiTaxId, 10, 64)
	if err != nil {
		tx.Rollback()
		return err
	}

	name := entry.Cluster.OrganismName

	tax_id, err := getOrCreateTaxId(name, ncbi_taxid, ctx, tx)
	if err != nil {
		switch {
		case errors.Is(err, errTaxidOutdated):
			ncbi_taxid = tax_id
			tax_id, err = getOrCreateTaxId(name, ncbi_taxid, ctx, tx)
			if err != nil {
				tx.Rollback()
				return err
			}
		default:
			tx.Rollback()
			return err
		}
	}

	statement := `INSERT INTO mibig.entries (
		entry_id, minimal, tax_id, organism_name, biosyn_class, legacy_comment
	) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	cluster := entry.Cluster

	args := []interface{}{
		cluster.MibigAccession,
		cluster.Minimal,
		tax_id,
		cluster.OrganismName,
		pq.Array(cluster.BiosyntheticClasses),
		entry.Comments,
	}

	_, err = tx.ExecContext(ctx, statement, args...)
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

var errTaxidOutdated = errors.New("taxId outdated, please retry")

func getOrCreateTaxId(name string, ncbi_taxid int64, ctx context.Context, tx *sql.Tx) (int64, error) {
	var tax_id int64

	args := []interface{}{
		ncbi_taxid,
		name,
	}

	err := tx.QueryRow(`SELECT tax_id FROM mibig.taxa WHERE ncbi_taxid = $1 AND name = $2`, args...).Scan(&tax_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ncbiTaxEntry, err := data.EntryForTaxId(ncbi_taxid)
			if err != nil {
				tx.Rollback()
				return -1, err
			}

			if ncbi_taxid != ncbiTaxEntry.TaxId {
				return ncbiTaxEntry.TaxId, errTaxidOutdated
			}

			query := `INSERT INTO mibig.taxa
				(ncbi_taxid, superkingdom, kingdom, phylum, class, taxonomic_order, family, genus, species, name)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING tax_id`

			args := []interface{}{
				ncbi_taxid,
				ncbiTaxEntry.Superkingdom,
				ncbiTaxEntry.Kingdom,
				ncbiTaxEntry.Phylum,
				ncbiTaxEntry.Class,
				ncbiTaxEntry.Order,
				ncbiTaxEntry.Family,
				ncbiTaxEntry.Genus,
				ncbiTaxEntry.Species,
				name,
			}

			err = tx.QueryRow(query, args...).Scan(&tax_id)
			if err != nil {
				tx.Rollback()
				return -1, err
			}

		} else {
			tx.Rollback()
			return -1, err
		}
	}
	return tax_id, nil
}

func resolveBgcSubclasses(entry data.MibigEntry) []string {
	resolved_types := []string{}

	for _, bgc_type := range entry.Cluster.BiosyntheticClasses {
		switch bgc_type {
		case "NRP":
			resolved_types = append(resolved_types, resolveNrpSubtype(entry)...)
		case "Polyketide":
			resolved_types = append(resolved_types, resolvePksSubtype(entry)...)
		case "Saccharide":
			resolved_types = append(resolved_types, resolveSaccharideSubtype(entry)...)
		case "RiPP":
			resolved_types = append(resolved_types, resolveRiPPSubtype(entry)...)
		case "Other":
			resolved_types = append(resolved_types, resolveOtherSubtype(entry)...)
		default:
			resolved_types = append(resolved_types, string(bgc_type))
		}
	}

	return resolved_types
}

func resolveNrpSubtype(entry data.MibigEntry) []string {
	return []string{"NRP"}
}

func resolvePksSubtype(entry data.MibigEntry) []string {
	subtypes := []string{}
	if reflect.ValueOf(entry.Cluster.Polyketide).IsNil() ||
		len(entry.Cluster.Polyketide.Synthases) < 1 {
		return []string{"Polyketide"}
	}
	for _, subtype := range entry.Cluster.Polyketide.Synthases {
		for _, subclass := range subtype.Subclass {
			subtypes = append(subtypes, fmt.Sprintf("%s polyketide", subclass))
		}
	}

	if len(subtypes) > 0 {
		return subtypes
	}
	return []string{"Polyketide"}
}

func resolveSaccharideSubtype(entry data.MibigEntry) []string {
	return []string{"Saccharide"}
}

func resolveRiPPSubtype(entry data.MibigEntry) []string {
	if reflect.ValueOf(entry.Cluster.RiPP).IsNil() {
		return []string{"RiPP"}
	}

	if entry.Cluster.RiPP.Subclass != "" {
		subclass := entry.Cluster.RiPP.Subclass
		if subclass == "Head-to-tailcyclized peptide" {
			// typoed and outdated name
			subclass = "Sactipeptide"
		}
		return []string{subclass}
	}
	return []string{"RiPP"}
}

func resolveOtherSubtype(entry data.MibigEntry) []string {
	return []string{"Other"}
}

func insertEntryBgcRelation(entry data.MibigEntry, ctx context.Context, tx *sql.Tx) error {
	for _, biosyn_class := range resolveBgcSubclasses(entry) {
		fmt.Printf("%s\n", biosyn_class)
		statement := `SELECT bgc_type_id FROM mibig.bgc_types WHERE name = $1;`
		var bgc_type_id int
		err := tx.QueryRowContext(ctx, statement, biosyn_class).Scan(&bgc_type_id)
		if err != nil {
			tx.Rollback()
			return err
		}

		statement = `INSERT INTO mibig.rel_entries_types (entry_id, bgc_type_id) VALUES ($1, $2);`
		_, err = tx.ExecContext(ctx, statement, entry.Cluster.MibigAccession, bgc_type_id)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return nil
}

func insertChangeLog(entry data.MibigEntry, ctx context.Context, tx *sql.Tx) error {
	for _, change_log := range entry.ChangeLogs {
		statement := `SELECT version_id FROM mibig.versions WHERE name = $1;`
		var version_id int
		err := tx.QueryRowContext(ctx, statement, change_log.Version).Scan(&version_id)
		if err != nil {
			tx.Rollback()
			fmt.Printf("missing version %s\n", change_log.Version)
			return err
		}

		statement = `INSERT INTO mibig.changelogs (
	comments, version_id, entry_id ) VALUES ($1, $2, $3);`
		args := []interface{}{
			pq.Array(change_log.Comments),
			version_id,
			entry.Cluster.MibigAccession,
		}

		_, err = tx.ExecContext(ctx, statement, args...)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return nil
}

func insertPublications(entry data.MibigEntry, ctx context.Context, tx *sql.Tx) error {
	for _, publication := range entry.Cluster.Publications {
		statement := `SELECT pub_id FROM mibig.publications WHERE pub_reference = $1;`
		var pub_id int

		err := tx.QueryRowContext(ctx, statement, publication.Reference).Scan(&pub_id)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				tx.Rollback()
				return err
			}
			statement = `INSERT INTO mibig.publications (pub_type, pub_reference)
						 VALUES ($1, $2) RETURNING pub_id;`
			err := tx.QueryRowContext(ctx, statement, publication.Type, publication.Reference).Scan(&pub_id)
			if err != nil {
				tx.Rollback()
				return err
			}
		}

		statement = `INSERT INTO mibig.rel_entries_pubs (entry_id, pub_id) VALUES ($1, $2);`
		_, err = tx.ExecContext(ctx, statement, entry.Cluster.MibigAccession, pub_id)
		if err != nil {
			tx.Rollback()
			return err
		}

	}
	return nil
}

func insertCompounds(entry data.MibigEntry, ctx context.Context, tx *sql.Tx) error {
	for _, compound := range entry.Cluster.Compounds {
		statement := `INSERT INTO mibig.chem_compounds (
			name, synonyms, entry_id
		) VALUES ($1, $2, $3);`

		_, err := tx.ExecContext(ctx, statement, compound.Name, pq.Array(compound.Synonyms),
			entry.Cluster.MibigAccession)
		if err != nil {
			tx.Rollback()
			return err
		}

	}
	return nil
}

func insertLocus(entry data.MibigEntry, ctx context.Context, tx *sql.Tx) error {
	statement := `INSERT INTO mibig.loci (
		accession, completeness, evidences, start_coord, end_coord, mixs_compliant, entry_id
	) VALUES ($1, $2, $3, $4, $5, $6, $7);`

	locus := entry.Cluster.Loci
	args := []interface{}{
		locus.Accession,
		strings.ToLower(locus.Completeness),
		pq.Array(locus.Evidence),
		locus.StartCoord,
		locus.EndCoord,
		locus.MixsCompliant,
		entry.Cluster.MibigAccession,
	}

	_, err := tx.ExecContext(ctx, statement, args...)
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
