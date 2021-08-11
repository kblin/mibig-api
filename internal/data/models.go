package data

import (
	"time"

	"secondarymetabolites.org/mibig-api/internal/queries"
)

type JsonData map[string]interface{}

type Entry struct {
	ID    int      `db:"entry_id"`
	Acc   string   `db:"acc"`
	TaxID int      `db:"tax_id"`
	Data  JsonData `db:"data"`
}

type StatCluster struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Count       int    `json:"count"`
	Class       string `json:"css_class"`
}

type StatCounts struct {
	Total      int `json:"total"`
	Minimal    int `json:"minimal"`
	Complete   int `json:"complete"`
	Incomplete int `json:"incomplete"`
}

type TaxonStats struct {
	Genus string `json:"genus"`
	Count int    `json:"count"`
}

type ProductTag struct {
	Name  string `json:"name"`
	Class string `json:"css_class"`
}

type Compound struct {
	Name string `json:"compound"`
}

type CompoundList []Compound

type Product struct {
	Name     string   `json:"name"`
	Synonyms []string `json:"synonyms,omitempty"`
}

type RepositoryEntry struct {
	Accession    string       `json:"accession"`
	Minimal      bool         `json:"minimal"`
	Complete     string       `json:"complete"`
	Products     []Product    `json:"products"`
	ProductTags  []ProductTag `json:"classes"`
	OrganismName string       `json:"organism"`
}

type LabelsAndCounts struct {
	Labels []string `json:"labels"`
	Data   []int    `json:"data"`
}

type ResultStats struct {
	ClustersByType   *LabelsAndCounts `json:"clusters_by_type"`
	ClustersByPhylun *LabelsAndCounts `json:"clusters_by_phylun"`
}

type AccessionRequestLocus struct {
	GenBankAccession string `json:"genbank_accession"`
	Start            int    `json:"start"`
	End              int    `json:"end"`
}

type AccessionRequest struct {
	Name      string                  `json:"name"`
	Email     string                  `json:"email"`
	Compounds []string                `json:"compounds"`
	Loci      []AccessionRequestLocus `json:"loci"`
}

type MibigModel interface {
	Counts() (*StatCounts, error)
	ClusterStats() ([]StatCluster, error)
	GenusStats() ([]TaxonStats, error)
	Repository() ([]RepositoryEntry, error)
	Search(t queries.QueryTerm) ([]string, error)
	Get(ids []string) ([]RepositoryEntry, error)
	Available(category string, term string) ([]AvailableTerm, error)
	ResultStats(ids []string) (*ResultStats, error)
	GuessCategories(query *queries.Query) error
	LookupContributors(ids []string) ([]Contributor, error)
}

type AvailableTerm struct {
	Val  string `json:"val"`
	Desc string `json:"desc"`
}

type LegacySubmission struct {
	Id        int
	Submitted time.Time
	Modified  time.Time
	Raw       string
	Version   int
}

type LegacyGeneSubmission struct {
	Id        int
	BgcId     string
	Submitted time.Time
	Modified  time.Time
	Raw       string
	Version   int
}

type LegacyNrpsSubmission struct {
	Id        int
	BgcId     string
	Submitted time.Time
	Modified  time.Time
	Raw       string
	Version   int
}

type LecagyModel interface {
	CreateSubmission(submission *LegacySubmission) error
	CreateGeneSubmission(submission *LegacyGeneSubmission) error
	CreateNrpsSubmission(submission *LegacyNrpsSubmission) error
}

type Contributor struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Organisation string `json:"organisation"`
}

type Submitter struct {
	Id           string
	Email        string
	Name         string
	CallName     string
	Institution  string
	PasswordHash []byte
	Public       bool
	GDPRConsent  bool
	Active       bool
	Roles        []Role
	Version      int
}

type Role struct {
	Id          int
	Name        string
	Description string
}

func RolesToStrings(roles []Role) []string {
	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}
	return roleNames
}
