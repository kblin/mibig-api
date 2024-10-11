package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"secondarymetabolites.org/mibig-api/internal/data"
)

func (m *LiveEntryModel) Add(entry data.MibigEntry, raw []byte, taxCache *data.TaxonCache) error {
	ctx := context.Background()
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = insertEntry(entry, taxCache, raw, ctx, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (m *LiveEntryModel) Refresh() error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err := m.DB.ExecContext(ctx, `REFRESH MATERIALIZED VIEW live.entry_bgc_info`)
	if err != nil {
		return err
	}
	_, err = m.DB.ExecContext(ctx, `REFRESH MATERIALIZED VIEW live.entry_compounds`)
	return err
}

func (m *LiveEntryModel) List() ([]data.MibigEntry, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	entries := []data.MibigEntry{}
	statement := `SELECT accession, version, status, quality, completeness, tax_id, organism_name, retirement_reason, see_also FROM live.entries`
	rows, err := m.DB.QueryContext(ctx, statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var entry data.MibigEntry
		err = rows.Scan(&entry.Accession, &entry.Version, &entry.Status, &entry.Quality, &entry.Completeness, &entry.Taxonomy.NcbiTaxId, &entry.Taxonomy.Name, pq.Array(&entry.RetirementReasons), pq.Array(&entry.SeeAlso))
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func (m *LiveEntryModel) Dump() error {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err := m.DB.ExecContext(ctx, `TRUNCATE live.entries CASCADE`)
	return err
}

func (m LiveEntryModel) LoadTaxonEntry(name string, ncbi_taxid int64, taxCache *data.TaxonCache) (int64, error) {
	ctx := context.Background()
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return -1, err
	}

	tax_id, err := getOrCreateTaxId(name, ncbi_taxid, taxCache, ctx, tx)
	if err != nil {
		switch {
		case errors.Is(err, errTaxidOutdated):
			for errors.Is(err, errTaxidOutdated) {
				ncbi_taxid = tax_id
				tax_id, err = getOrCreateTaxId(name, ncbi_taxid, taxCache, ctx, tx)
			}
			if err != nil {
				tx.Rollback()
				return -1, err
			}
		default:
			tx.Rollback()
			return -1, err
		}
	}

	return ncbi_taxid, tx.Commit()
}

func insertEntry(entry data.MibigEntry, taxCache *data.TaxonCache, raw []byte, ctx context.Context, tx *sql.Tx) error {

	statement := `INSERT INTO live.entries (
		entry_id, accession, version, status, quality, completeness, tax_id, organism_name, retirement_reason, see_also, data
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	tax_id, err := getOrCreateTaxId(entry.Taxonomy.Name, entry.Taxonomy.NcbiTaxId, taxCache, ctx, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	args := []interface{}{
		fmt.Sprintf("%s.%d", entry.Accession, entry.Version),
		entry.Accession,
		entry.Version,
		entry.Status,
		entry.Quality,
		entry.Completeness,
		tax_id,
		entry.Taxonomy.Name,
		pq.Array(entry.RetirementReasons),
		pq.Array(entry.SeeAlso),
		raw,
	}

	_, err = tx.ExecContext(ctx, statement, args...)
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

var errTaxidOutdated = errors.New("taxId outdated, please retry")

func getOrCreateTaxId(name string, ncbi_taxid int64, taxCache *data.TaxonCache, ctx context.Context, tx *sql.Tx) (int64, error) {
	var tax_id int64

	args := []interface{}{
		ncbi_taxid,
		name,
	}

	err := tx.QueryRow(`SELECT tax_id FROM data.taxa WHERE ncbi_taxid = $1 AND name = $2`, args...).Scan(&tax_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ncbiTaxEntry, err := taxCache.EntryForTaxId(ncbi_taxid)
			if err != nil {
				tx.Rollback()
				return -1, err
			}

			if ncbi_taxid != ncbiTaxEntry.TaxId {
				return ncbiTaxEntry.TaxId, errTaxidOutdated
			}

			query := `INSERT INTO data.taxa
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
