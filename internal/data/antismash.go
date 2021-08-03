package data

import "encoding/json"

type AsEntry struct {
	Version   string     `json:"version"`
	InputFile string     `json:"input_file"`
	Taxon     string     `json:"taxon"`
	Timings   Timings    `json:"timings"`
	Records   []AsRecord `json:"records"`
	Schema    int64      `json:"schema"`
}

type Timings map[string]map[string]float64

type AsRecord struct {
	Id          string         `json:"id"`
	Annotations GbkAnnotations `json:"annotations"`
	Modules     AsModules      `json:"modules"`
}

type GbkAnnotations struct {
	Accessions []string `json:"accessions"`
	Organism   string   `json:"organism"`
	Taxonomy   []string `json:"taxonomy"`
}

type AsModules struct {
	HmmDetection AsHmmDetectionResults `json:"antismash.detection.hmm_detection"`
}

type AsHmmDetectionResults struct {
	RecordId      string        `json:"record_id"`
	EnabledTypes  []string      `json:"enabled_types"`
	SchemaVersion int64         `json:"schema_version"`
	RuleResults   AsRuleResults `json:"rule_results"`
}

type AsRuleResults struct {
	SchemaVersion     int64               `json:"schema_version"`
	Tool              string              `json:"tool"`
	CdsByProtocluster []CdsByProtocluster `json:"cds_by_protocluster"`
}

type CdsByProtocluster struct {
	Protocluster Protocluster
	Cdses        []ProtoclusterCds
}

func (cbp *CdsByProtocluster) UnmarshalJSON(bs []byte) error {
	arr := []json.RawMessage{}
	err := json.Unmarshal(bs, &arr)
	if err != nil {
		return err
	}

	pc := Protocluster{}
	cdses := make([]ProtoclusterCds, 5)

	err = json.Unmarshal(arr[0], &pc)
	if err != nil {
		return err
	}
	cbp.Protocluster = pc
	err = json.Unmarshal(arr[1], &cdses)
	if err != nil {
		return err
	}
	cbp.Cdses = cdses

	return nil
}

func (cbp *CdsByProtocluster) MarshalJSON() ([]byte, error) {
	arr := []interface{}{cbp.Protocluster, cbp.Cdses}
	return json.Marshal(&arr)
}

type Protocluster struct {
	Location   string              `json:"location"`
	Type       string              `json:"type"`
	Qualifiers map[string][]string `json:"qualifiers"` // TODO: Do we want to properly type this?
}

type ProtoclusterCds struct {
	Name              string              `json:"cds_name"`
	Domains           []CdsDomain         `json:"domains"`
	DefinitionDomains map[string][]string `json:"definition_domains"`
}

type CdsDomain struct {
	Name     string
	EValue   float64
	Bitscore float64
	NumSeeds float64 // Technicall an int, but JSON just has floats, so whatever.
	Tool     string
}

func (cd *CdsDomain) UnmarshalJSON(bs []byte) error {
	arr := []interface{}{}
	err := json.Unmarshal(bs, &arr)
	if err != nil {
		return err
	}

	cd.Name = arr[0].(string)
	cd.EValue = arr[1].(float64)
	cd.Bitscore = arr[2].(float64)
	cd.NumSeeds = arr[3].(float64)
	cd.Tool = arr[4].(string)

	return nil
}

func (cd *CdsDomain) MarshalJSON() ([]byte, error) {
	arr := []interface{}{cd.Name, cd.EValue, cd.Bitscore, cd.NumSeeds, cd.Tool}
	return json.Marshal(&arr)
}

type AsGeneFunctions struct {
}
