package data

type MibigTaxonomy struct {
	NcbiTaxId int64  `json:"ncbiTaxId"`
	Name      string `json:"name"`
}

type MibigEntry struct {
	Accession         string        `json:"accession"`
	Version           int           `json:"version"`
	Status            string        `json:"status"`
	Quality           string        `json:"quality"`
	Completeness      string        `json:"completeness"`
	Taxonomy          MibigTaxonomy `json:"taxonomy"`
	RetirementReasons []string      `json:"retirement_reasons,omitempty"`
	SeeAlso           []string      `json:"see_also,omitempty"`
	Comment           string        `json:"comment,omitempty"`
}
