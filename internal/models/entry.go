package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"secondarymetabolites.org/mibig-api/internal/data"
	"secondarymetabolites.org/mibig-api/internal/queries"
	"secondarymetabolites.org/mibig-api/internal/utils"
)

type EntryModel interface {
	Counts() (*data.StatCounts, error)
	ClusterStats() ([]data.StatCluster, error)
	GenusStats() ([]data.TaxonStats, error)
	Repository() ([]data.RepositoryEntry, error)
	Get(ids []string) ([]data.RepositoryEntry, error)
	Search(t queries.QueryTerm) ([]string, error)
	Available(category string, term string) ([]data.AvailableTerm, error)
	ResultStats(ids []string) (*data.ResultStats, error)
	GuessCategories(query *queries.Query) error
	LookupContributors(ids []string) ([]data.Contributor, error)

	Add(entry data.MibigEntry) error
	InsertEntryStatus(status data.MibigEntryStatus) error
}

type LiveEntryModel struct {
	DB *sql.DB
}

func NewEntryModel(db *sql.DB) *LiveEntryModel {
	return &LiveEntryModel{DB: db}
}

func (m *LiveEntryModel) Counts() (*data.StatCounts, error) {
	stmt_total := `SELECT COUNT(entry_id) FROM mibig.entries`
	stmt_minimal := `SELECT COUNT(entry_id) FROM mibig.entries WHERE minimal = TRUE`
	stmt_complete := `SELECT COUNT(entry_id) FROM mibig.entries LEFT JOIN mibig.loci USING (entry_id) WHERE completeness = 'complete'`
	stmt_incomplete := `SELECT COUNT(entry_id) FROM mibig.entries LEFT JOIN mibig.loci USING (entry_id) WHERE completeness = 'incomplete'`
	var counts data.StatCounts

	err := m.DB.QueryRow(stmt_total).Scan(&counts.Total)
	if err != nil {
		return nil, err
	}

	err = m.DB.QueryRow(stmt_minimal).Scan(&counts.Minimal)
	if err != nil {
		return nil, err
	}

	err = m.DB.QueryRow(stmt_complete).Scan(&counts.Complete)
	if err != nil {
		return nil, err
	}

	err = m.DB.QueryRow(stmt_incomplete).Scan(&counts.Incomplete)
	if err != nil {
		return nil, err
	}

	return &counts, nil
}

func (m *LiveEntryModel) ClusterStats() ([]data.StatCluster, error) {
	statement := `SELECT term, description, entry_count, safe_class FROM
	(
		SELECT unnest(biosyn_class) AS name, COUNT(1) AS entry_count FROM mibig.entries GROUP BY name
	) counter
	LEFT JOIN mibig.bgc_types t USING (name)
	ORDER BY entry_count DESC`

	var clusters []data.StatCluster

	rows, err := m.DB.Query(statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		cluster := data.StatCluster{}
		if err = rows.Scan(&cluster.Type, &cluster.Description, &cluster.Count, &cluster.Class); err != nil {
			return nil, err
		}
		clusters = append(clusters, cluster)
	}

	return clusters, nil
}

func (m *LiveEntryModel) GenusStats() ([]data.TaxonStats, error) {
	statement := `SELECT genus, COUNT(genus) AS ct FROM mibig.entries LEFT JOIN mibig.taxa USING (tax_id) GROUP BY genus ORDER BY ct DESC, genus`
	var stats []data.TaxonStats

	rows, err := m.DB.Query(statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		stat := data.TaxonStats{}
		if err = rows.Scan(&stat.Genus, &stat.Count); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (m *LiveEntryModel) Repository() ([]data.RepositoryEntry, error) {
	statement := `SELECT
		a.entry_id,
		a.minimal,
		l.completeness AS complete,
		array_agg(DISTINCT c.name) AS compounds,
		array_agg(array_to_json(synonyms)) AS synonyms,
		array_agg(DISTINCT safe_class || ':' || b.name ORDER BY safe_class || ':' || b.name) AS biosyn_class,
		t.name
	FROM mibig.entries a
	JOIN mibig.rel_entries_types USING (entry_id)
	JOIN mibig.bgc_types b USING (bgc_type_id)
	JOIN mibig.taxa t USING (tax_id)
	JOIN mibig.loci l USING (entry_id)
	JOIN mibig.chem_compounds c USING (entry_id)
	GROUP BY entry_id, minimal, complete, t.name
	ORDER BY entry_id`

	rows, err := m.DB.Query(statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return parseRepositoryEntriesFromDB(rows)
}

func parseRepositoryEntriesFromDB(rows *sql.Rows) ([]data.RepositoryEntry, error) {
	var entries []data.RepositoryEntry

	for rows.Next() {
		var (
			class_and_css    []string
			product_names    []string
			product_synonyms []sql.NullString
			products         []data.Product
		)

		entry := data.RepositoryEntry{}
		if err := rows.Scan(&entry.Accession, &entry.Minimal, &entry.Complete, pq.Array(&product_names),
			pq.Array(&product_synonyms), pq.Array(&class_and_css), &entry.OrganismName); err != nil {
			return nil, err
		}

		products = make([]data.Product, len(product_names))
		for i := range product_names {
			products[i] = data.Product{Name: product_names[i]}
			if !product_synonyms[i].Valid {
				continue
			}
			var parsed_synonyms []string
			err := json.Unmarshal([]byte(product_synonyms[i].String), &parsed_synonyms)
			if err != nil {
				return nil, err
			}
			products[i].Synonyms = parsed_synonyms
		}
		entry.Products = products

		for _, combined := range class_and_css {
			parts := strings.SplitN(combined, ":", 2)
			tag := data.ProductTag{Name: parts[1], Class: parts[0]}
			entry.ProductTags = append(entry.ProductTags, tag)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (m *LiveEntryModel) Get(ids []string) ([]data.RepositoryEntry, error) {
	statement := `SELECT
		a.entry_id,
		a.minimal,
		l.completeness AS complete,
		array_cat(array_agg(DISTINCT c.name), synonyms) AS compounds,
		array_agg(DISTINCT safe_class || ':' || b.name ORDER BY safe_class || ':' || b.name) AS biosyn_class,
		t.name
	FROM ( SELECT * FROM unnest($1::text[]) AS entry_id) vals
	JOIN mibig.entries a USING (entry_id)
	JOIN mibig.rel_entries_types USING (entry_id)
	JOIN mibig.bgc_types b USING (bgc_type_id)
	JOIN mibig.chem_compounds c USING (entry_id)
	JOIN mibig.taxa t USING (tax_id)
	JOIN mibig.loci l USING (entry_id)
	GROUP BY a.entry_id, minimal, complete, synonyms, t.name
	ORDER BY a.entry_id`

	rows, err := m.DB.Query(statement, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return parseRepositoryEntriesFromDB(rows)
}

var categoryDetector = map[string]string{
	"type":     `SELECT COUNT(bgc_type_id) FROM mibig.bgc_types WHERE term ILIKE $1`,
	"acc":      `SELECT COUNT(entry_id) FROM mibig.entries WHERE entry_id ILIKE $1`,
	"compound": `SELECT COUNT(entry_id) FROM mibig.chem_compounds WHERE name ILIKE $1`,
	"genus":    `SELECT COUNT(tax_id) FROM mibig.taxa WHERE genus ILIKE $1`,
	"species":  `SELECT COUNT(tax_id) FROM mibig.taxa WHERE species ILIKE $1`,
}

func (m *LiveEntryModel) guessCategory(term string) (string, error) {

	for _, category := range []string{"type", "acc", "compound", "genus", "species"} {
		statement := categoryDetector[category]
		var count int
		if err := m.DB.QueryRow(statement, term).Scan(&count); err != nil {
			return "", err
		}
		if count > 0 {
			return category, nil
		}
	}
	return "", data.ErrInvalidCategory
}

var statementByCategory = map[string]string{
	"type": `SELECT entry_id FROM mibig.entries e LEFT JOIN mibig.rel_entries_types ret USING (entry_id) WHERE bgc_type_id IN (
	WITH RECURSIVE all_subtypes AS (
		SELECT bgc_type_id, parent_id FROM mibig.bgc_types WHERE term = $1
	UNION
		SELECT r.bgc_type_id, r.parent_id FROM mibig.bgc_types r INNER JOIN all_subtypes s ON s.bgc_type_id = r.parent_id)
	SELECT bgc_type_id FROM all_subtypes)`,
	"compound":     `SELECT entry_id FROM mibig.compounds WHERE name ILIKE $1`,
	"acc":          `SELECT entry_id FROM mibig.entries WHERE entry_id ILIKE $1`,
	"superkingdom": `SELECT entry_id FROM mibig.entries LEFT JOIN mibig.taxa USING (tax_id) WHERE superkingdom ILIKE $1`,
	"kingdom":      `SELECT entry_id FROM mibig.entries LEFT JOIN mibig.taxa USING (tax_id) WHERE kingdom ILIKE $1`,
	"phylum":       `SELECT entry_id FROM mibig.entries LEFT JOIN mibig.taxa USING (tax_id) WHERE phylum ILIKE $1`,
	"class":        `SELECT entry_id FROM mibig.entries LEFT JOIN mibig.taxa USING (tax_id) WHERE class ILIKE $1`,
	"order":        `SELECT entry_id FROM mibig.entries LEFT JOIN mibig.taxa USING (tax_id) WHERE taxonomic_order ILIKE $1`,
	"family":       `SELECT entry_id FROM mibig.entries LEFT JOIN mibig.taxa USING (tax_id) WHERE family ILIKE $1`,
	"genus":        `SELECT entry_id FROM mibig.entries LEFT JOIN mibig.taxa USING (tax_id) WHERE genus ILIKE $1`,
	"species":      `SELECT entry_id FROM mibig.entries LEFT JOIN mibig.taxa USING (tax_id) WHERE species ILIKE $1`,
	"minimal":      `SELECT entry_id FROM mibig.entries WHERE minimal = $1`,
	"completeness": `SELECT entry_id FROM mibig.entries LEFT JOIN mibig.loci USING (entry_id) WHERE completeness = $1`,
	"ncbi":         `SELECT entry_id FROM mibig.entries LEFT JOIN mibig.loci USING (entry_id) WHERE accession ILIKE $1`,
}

func (m *LiveEntryModel) Search(t queries.QueryTerm) ([]string, error) {
	var entry_ids []string
	switch v := t.(type) {
	case *queries.Expression:
		if v.Category == "unknown" {
			cat, err := m.guessCategory(v.Term)
			if err != nil {
				return nil, err
			}
			v.Category = cat
		}
		statement, ok := statementByCategory[v.Category]
		if !ok {
			return []string{}, nil
		}

		rows, err := m.DB.Query(statement, v.Term)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var entry_id string
			rows.Scan(&entry_id)
			entry_ids = append(entry_ids, entry_id)
		}

		return entry_ids, nil

	case *queries.Operation:
		var (
			err   error
			left  []string
			right []string
		)
		left, err = m.Search(v.Left)
		if err != nil {
			return nil, err
		}
		right, err = m.Search(v.Right)
		if err != nil {
			return nil, err
		}
		switch v.Operation {
		case queries.AND:
			return utils.IntersectString(left, right), nil
		case queries.OR:
			return utils.UnionString(left, right), nil
		case queries.EXCEPT:
			return utils.DifferenceString(left, right), nil
		default:
			return nil, fmt.Errorf("invalid operation: %s", v.Op())
		}
	}
	// Should never get here
	return entry_ids, nil
}

var availableByCategory = map[string]string{
	"type":         `SELECT DISTINCT(term), description FROM mibig.bgc_types WHERE term ILIKE concat($1::text, '%') OR description ILIKE concat($1::text, '%') ORDER BY term`,
	"compound":     `SELECT DISTINCT(name), name FROM mibig.compounds WHERE name ILIKE concat($1::text, '%')`,
	"acc":          `SELECT DISTINCT(entry_id), entry_id FROM mibig.entries WHERE entry_id ILIKE concat('%', $1::text, '%')`,
	"superkingdom": `SELECT DISTINCT(superkingdom), superkingdom FROM mibig.taxa WHERE superkingdom ILIKE concat('%', $1::text, '%')`,
	"kingdom":      `SELECT DISTINCT(kingdom), kingdom FROM mibig.taxa WHERE kingdom ILIKE concat('%', $1::text, '%')`,
	"phylum":       `SELECT DISTINCT(phylum), phylum FROM mibig.taxa WHERE phylum ILIKE concat('%', $1::text, '%')`,
	"class":        `SELECT DISTINCT(class), class FROM mibig.taxa WHERE class ILIKE concat('%', $1::text, '%')`,
	"order":        `SELECT DISTINCT(taxonomic_order), taxonomic_order FROM mibig.taxa WHERE taxonomic_order ILIKE concat('%', $1::text, '%')`,
	"family":       `SELECT DISTINCT(family), family FROM mibig.taxa WHERE family ILIKE concat('%', $1::text, '%')`,
	"genus":        `SELECT DISTINCT(genus), genus FROM mibig.taxa WHERE genus ILIKE concat('%', $1::text, '%')`,
	"species":      `SELECT DISTINCT(species), species FROM mibig.taxa WHERE species ILIKE concat('%', $1::text, '%')`,
	"completeness": `WITH completeness AS (SELECT unnest(enum_range(NULL::mibig.locus_completeness))::text AS value) SELECT value FROM completeness WHERE value ILIKE concat($1::text, '%')`,
	"ncbi":         `SELECT DISTINCT(accession), accession FROM mibig.loci WHERE accession ILIKE concat($1::text, '%')`,
}

func (m *LiveEntryModel) Available(category string, term string) ([]data.AvailableTerm, error) {
	var (
		available []data.AvailableTerm
		statement string
		ok        bool
	)

	if category == "minimal" {
		description := "Minimal MIBiG entry"
		return fakeBooleanOptions(term, description)
	}

	if statement, ok = availableByCategory[category]; !ok {
		return nil, data.ErrInvalidCategory
	}
	rows, err := m.DB.Query(statement, term)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var av data.AvailableTerm
		err = rows.Scan(&av.Val, &av.Desc)
		if err != nil {
			return nil, err
		}
		available = append(available, av)
	}
	return available, nil
}

func fakeBooleanOptions(term string, description string) ([]data.AvailableTerm, error) {
	if strings.HasPrefix("true", term) {
		return []data.AvailableTerm{{Val: "true", Desc: description}}, nil
	}
	if strings.HasPrefix("yes", term) {
		return []data.AvailableTerm{{Val: "true", Desc: description}}, nil
	}
	if strings.HasPrefix("false", term) {
		return []data.AvailableTerm{{Val: "false", Desc: description}}, nil
	}
	if strings.HasPrefix("no", term) {
		return []data.AvailableTerm{{Val: "false", Desc: description}}, nil
	}
	return []data.AvailableTerm{}, nil
}

func (m *LiveEntryModel) ResultStats(ids []string) (*data.ResultStats, error) {
	var stats data.ResultStats
	var err error

	cluster_by_type_search := `SELECT
	unnest(biosyn_class) as biosyn_class, COUNT(1) AS class_count
	FROM ( SELECT * FROM unnest($1::text[]) AS entry_id) vals
	JOIN mibig.entries a USING (entry_id)
	GROUP BY biosyn_class`

	cluster_by_phylum_search := `SELECT
	phylum, COUNT(phylum)
	FROM ( SELECT * FROM unnest($1::text[]) AS entry_id) vals
	JOIN mibig.entries USING (entry_id)
	JOIN mibig.taxa USING (tax_id)
	GROUP BY phylum`

	stats.ClustersByType, err = m.labelsAndCounts(cluster_by_type_search, ids)
	if err != nil {
		return nil, err
	}

	stats.ClustersByPhylun, err = m.labelsAndCounts(cluster_by_phylum_search, ids)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (m *LiveEntryModel) labelsAndCounts(statement string, ids []string) (*data.LabelsAndCounts, error) {
	var lc data.LabelsAndCounts

	rows, err := m.DB.Query(statement, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var label string
		var count int
		err = rows.Scan(&label, &count)
		if err != nil {
			return nil, err
		}
		lc.Labels = append(lc.Labels, label)
		lc.Data = append(lc.Data, count)
	}
	return &lc, nil
}

func (m *LiveEntryModel) GuessCategories(query *queries.Query) error {
	return m.recursiveGuessCategories(query.Terms)
}

func (m *LiveEntryModel) recursiveGuessCategories(term queries.QueryTerm) error {
	switch v := term.(type) {
	case *queries.Expression:
		if v.Category == "unknown" {
			cat, err := m.guessCategory(v.Term)
			if err != nil {
				return err
			}
			v.Category = cat
		}
	case *queries.Operation:
		if err := m.recursiveGuessCategories(v.Left); err != nil {
			return err
		}
		if err := m.recursiveGuessCategories(v.Right); err != nil {
			return err
		}
	}
	return nil
}

func (m *LiveEntryModel) LookupContributors(ids []string) ([]data.Contributor, error) {
	statement := `SELECT a.user_id, name, email, institution
	FROM ( SELECT * FROM unnest($1::text[]) AS user_id) vals
	JOIN mibig_submitters.submitters a USING (user_id)
	WHERE is_public = TRUE AND gdpr_consent = TRUE;
	`
	rows, err := m.DB.Query(statement, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contributors []data.Contributor

	for rows.Next() {
		contributor := data.Contributor{}
		err = rows.Scan(&contributor.Id, &contributor.Name, &contributor.Email, &contributor.Organisation)
		if err != nil {
			return nil, err
		}
		contributors = append(contributors, contributor)
	}
	return contributors, nil
}

/* type EntryModel interface {
	Counts() (*data.StatCounts, error)
	ClusterStats() ([]data.StatCluster, error)
	GenusStats() ([]data.TaxonStats, error)
	Repository() ([]data.RepositoryEntry, error)
	Get(ids []string) ([]data.RepositoryEntry, error)
	Search(t queries.QueryTerm) ([]string, error)
	Available(category string, term string) ([]data.AvailableTerm, error)
	ResultStats(ids []string) (*data.ResultStats, error)
	GuessCategories(query *queries.Query) error
	LookupContributors(ids []string) ([]data.Contributor, error)
} */

type MockEntryModel struct {
}

func NewMockEntryModel() *MockEntryModel {
	return &MockEntryModel{}
}

func (m *MockEntryModel) Counts() (*data.StatCounts, error) {
	return nil, data.ErrNotImplemented
}

func (m *MockEntryModel) ClusterStats() ([]data.StatCluster, error) {
	return nil, data.ErrNotImplemented
}

func (m *MockEntryModel) GenusStats() ([]data.TaxonStats, error) {
	return nil, data.ErrNotImplemented
}

func (m *MockEntryModel) Repository() ([]data.RepositoryEntry, error) {
	return nil, data.ErrNotImplemented
}

func (m *MockEntryModel) Get(ids []string) ([]data.RepositoryEntry, error) {
	return nil, data.ErrNotImplemented
}

func (m *MockEntryModel) Search(t queries.QueryTerm) ([]string, error) {
	return nil, data.ErrNotImplemented
}

func (m *MockEntryModel) Available(category string, term string) ([]data.AvailableTerm, error) {
	return nil, data.ErrNotImplemented
}

func (m *MockEntryModel) ResultStats(ids []string) (*data.ResultStats, error) {
	return nil, data.ErrNotImplemented
}

func (m *MockEntryModel) GuessCategories(query *queries.Query) error {
	return data.ErrNotImplemented
}

func (m *MockEntryModel) LookupContributors(ids []string) ([]data.Contributor, error) {
	return nil, data.ErrNotImplemented
}

func (m *MockEntryModel) Add(entry data.MibigEntry) error {
	return data.ErrNotImplemented
}

func (m *MockEntryModel) InsertEntryStatus(status data.MibigEntryStatus) error {
	return data.ErrNotImplemented
}
