package data

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type ChangeLog struct {
	Comments     []string `json:"comments"`
	Contributors []string `json:"contributors"`
	Version      string   `json:"version"`
}

type BiosyntheticClass string

const (
	BiosynClassAlkaloid   BiosyntheticClass = "Alkaloid"
	BiosynClassNrp        BiosyntheticClass = "NRP"
	BiosynClassPolyketide BiosyntheticClass = "Polyketide"
	BiosynClassRipp       BiosyntheticClass = "RiPP"
	BiosynClassSaccharide BiosyntheticClass = "Saccharide"
	BiosynClassTerpene    BiosyntheticClass = "Terpene"
	BiosynClassOther      BiosyntheticClass = "Other"
)

func (bc *BiosyntheticClass) UnmarshalJSON(b []byte) error {
	type BC BiosyntheticClass
	var s *BC = (*BC)(bc)
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	switch *bc {
	case BiosynClassAlkaloid,
		BiosynClassNrp,
		BiosynClassPolyketide,
		BiosynClassRipp,
		BiosynClassSaccharide,
		BiosynClassTerpene,
		BiosynClassOther:
		return nil
	}

	return errors.New("invalid BiosyntheticClass type")
}

type ChemTarget struct {
	Name         string   `json:"target"`
	Publications []string `json:"publications,omitempty"`
}

type ChemMoiety struct {
	Name       string   `json:"moiety"`
	Subcluster []string `json:"subcluster,omitempty"` // TODO: link to gene names from cluster?
}

// Sorted in order of the appearance in the JSON spec
type ChemCompound struct {
	ChemActivities   []string     `json:"chem_acts,omitempty"`
	ChemMoieties     []ChemMoiety `json:"chem_moieties,omitempty"`
	ChemStructure    string       `json:"chem_struct,omitempty"`
	Synonyms         []string     `json:"chem_synonyms,omitempty"`
	ChemTargets      []ChemTarget `json:"chem_targets,omitempty"`
	Name             string       `json:"compound"`
	DatabaseIds      []string     `json:"database_id,omitempty"`
	Evidence         []string     `json:"evidence,omitempty"`
	MassSpecIonType  string       `json:"mass_spec_ion_type,omitempty"`
	MolecularMass    float64      `json:"mol_mass,omitempty"`
	MolecularFormula string       `json:"molecular_formula,omitempty"`
}

type Loci struct {
	Accession     string   `json:"accession"`
	Completeness  string   `json:"completeness"` // TODO: Turn into an enum
	EndCoord      int64    `json:"end_coord,omitempty"`
	Evidence      []string `json:"evidence,omitempty"` // TODO: Turn into an enum
	MixsCompliant bool     `json:"mixs_compliant,omitempty"`
	StartCoord    int64    `json:"start_coord,omitempty"`
}

type Genes struct {
	Annotations []GeneAnnotations `json:"annotations,omitempty"`
	ExtraGenes  []ExtraGene       `json:"extra_genes,omitempty"`
	Operons     []Operon          `json:"operons,omitempty"`
}

type GeneAnnotations struct {
	Comments          string          `json:"comments,omitempty"`
	Functions         []GeneFunctions `json:"functions,omitempty"`
	Id                string          `json:"id"`
	MutationPhenotype string          `json:"mut_pheno,omitempty"`
	Name              string          `json:"name,omitempty"`
	Product           string          `json:"product,omitempty"`
	Publications      []string        `json:"publications,omitempty"`
	Tailoring         []string        `json:"tailoring,omitempty"`
}

type GeneFunctions struct {
	Category string   `json:"category"` // TODO: restrict possible categories?
	Evidence []string `json:"evidence"`
}

type ExtraGene struct {
	Id          string    `json:"id"`
	Location    *Location `json:"location,omitempty"`
	Translation string    `json:"translation,omitempty"`
}

type Location struct {
	Exons  []Exon `json:"exons"`
	Strand int    `json:"strand"` // TODO: Turn into an enum
}

type Exon struct {
	End   int `json:"end"`
	Start int `json:"start"`
}

type Operon struct {
	Evidence []string `json:"evidence"` // TODO: Turn into an enum
	Genes    []string `json:"genes"`    // TODO: Validate against the available genes
}

type Thioesterase struct {
	Gene string `json:"gene"`
	Type string `json:"thioesterase_type"` // TODO: Turn into an enum // TODO: rename in JSON schema?
}

type ADomainSubstrateSpecificity struct {
	AminoAcidSubcluster []string `json:"aa_subcluster,omitempty"`
	Epimerized          bool     `json:"epimerized"`
	Evidence            []string `json:"evidence,omitempty"`
	Nonproteinogenic    []string `json:"nonproteinogenic,omitempty"`
	Proteinogenic       []string `json:"proteinogenic,omitempty"`
}

type NonCanonicalModule struct {
	Evidence      []string `json:"evidence,omitempty"`
	Iterated      bool     `json:"iterated"`
	NonElongating bool     `json:"non_elongating"`
	Skipped       bool     `json:"skipped"`
}

type NrpsModule struct {
	ADomainSubstrateSpecificity *ADomainSubstrateSpecificity `json:"a_substr_spec,omitempty"`
	Active                      bool                         `json:"active"`
	CDomainSubtype              string                       `json:"c_dom_subtype,omitempty"`
	Comments                    string                       `json:"comments,omitempty"`
	ModificationDomains         []string                     `json:"modification_domains,omitempty"`
	ModuleNumber                string                       `json:"module_number,omitempty"`
	NonCanonicalModule          *NonCanonicalModule          `json:"non_canonical,omitempty"`
}

type NrpsGene struct {
	Id      string       `json:"gene_id"`
	Modules []NrpsModule `json:"modules,omitempty"`
}

type PksIterative struct {
	CyclizationType string   `json:"cyclization_type"`
	Evidence        []string `json:"evidence,omitempty"`
	Genes           []string `json:"genes,omitempty"` // TODO: Validate against the synthase entry's gene list
	NrIterations    int64    `json:"nr_iterations,omitempty"`
	Subtype         string   `json:"subtype"` // TODO: Turn into an enum
}

type PksModule struct {
	ATSpecificities     []string            `json:"at_specificities"`
	Comments            string              `json:"comments,omitempty"`
	Domains             []string            `json:"domains,omitempty"`
	Evidence            string              `json:"evidence,omitempty"`      // TODO: Should be a list in the spec?
	Genes               []string            `json:"genes,omitempty"`         // TODO: Validate against the synthase entry's gene list
	KRStereoChemistry   string              `json:"kr_stereochem,omitempty"` // TODO: Turn into an enum
	ModuleNumber        string              `json:"module_number,omitempty"`
	NonCanonicalModule  *NonCanonicalModule `json:"non_canonical,omitempty"`
	ModificationDomains []string            `json:"pks_mod_doms,omitempty"`
}

type TransAT struct {
	Genes []string `json:"genes,omitempty"` // TODO: Validate against the synthase entry's gene list
}

type PksSynthase struct {
	Genes                   []string       `json:"genes"`
	Iterative               *PksIterative  `json:"iterative,omitempty"`
	Modules                 []PksModule    `json:"modules,omitempty"`
	PUFAModificationDomains []string       `json:"pufa_modification_domains,omitempty"`
	Subclass                []string       `json:"subclass,omitempty"`
	Thioesterases           []Thioesterase `json:"thioesterases,omitempty"`
	TransAT                 *TransAT       `json:"trans_at,omitempty"`
}

type RippPrecursor struct {
	CleavageRecognitionSites []string        `json:"cleavage_recogn_site,omitempty"` // TODO: Turn into plural in spec?
	CoreSequences            []string        `json:"core_sequence"`
	Crosslinks               []RippCrosslink `json:"crosslinks,omitempty"`
	FollowerSequence         string          `json:"follower_sequence,omitempty"`
	GeneId                   string          `json:"gene_id"` // TODO: Validate against cluster's gene list
	LeaderSequence           string          `json:"leader_sequence,omitempty"`
	RecognitionMotif         string          `json:"recognition_motif,omitempty"`
}

type RippCrosslink struct {
	Type     string `json:"crosslink_type"`
	FirstAA  int64  `json:"first_AA,omitempty"`
	SecondAA int64  `json:"second_AA,omitempty"`
}

type Glycosyltransferase struct {
	Evidence    []string `json:"evidence"` // TODO: Turn into an enum
	GeneId      string   `json:"gene_id"`  // TODO: Valdiate against cluster's gene list
	Specificity string   `json:"specificity"`
}

type Alkaloid struct {
	Subclass string `json:"subclass,omitempty"`
}

type Nrp struct {
	Cyclic        bool           `json:"cyclic"`
	LipidMoiety   string         `json:"lipid_moiety,omitempty"`
	NrpsGenes     []NrpsGene     `json:"nrps_genes,omitempty"`
	ReleaseType   []string       `json:"release_type,omitempty"`
	Subclass      string         `json:"subclass,omitempty"`
	Thioesterases []Thioesterase `json:"thioesterases,omitempty"`
}

type Other struct {
	Subclass string `json:"subclass"`
}

type Polyketide struct {
	Cyclases     []string      `json:"cyclases,omitempty"`
	Cyclic       bool          `json:"cyclic"`
	KetideLength int64         `json:"ketide_length,omitempty"`
	ReleaseType  []string      `json:"release_type,omitempty"`
	StarterUnit  string        `json:"starter_unit,omitempty"`
	Subclasses   []string      `json:"subclasses,omitempty"`
	Synthases    []PksSynthase `json:"synthases,omitempty"`
}

type RiPP struct {
	Cyclic     bool            `json:"cyclic"`
	Peptidases []string        `json:"peptidases,omitempty"` // TODO: Validate against cluster's gene list
	Precursors []RippPrecursor `json:"precursor_genes,omitempty"`
	Subclass   string          `json:"subclass,omitempty"`
}

type Saccharide struct {
	Glycosyltransferases []Glycosyltransferase `json:"glycosyltransferases,omitempty"`
	Subclass             string                `json:"subclass,omitempty"`
	SugarSubclusters     [][]string            `json:"sugar_subclusters,omitempty"`
}

type Terpene struct {
	CarbonCountSubclass      string   `json:"carbon_count_subclass,omitempty"`
	Prenyltransferases       []string `json:"prenyltransferases,omitempty"`
	StructuralSubclass       string   `json:"structural_subclass,omitempty"`
	TerpenePrecursor         string   `json:"terpene_precursor,omitempty"`
	TerpeneSynthasesCyclases []string `json:"terpene_synth_cycl,omitempty"`
}

type Publication struct {
	Type      string
	Reference string
}

func (p *Publication) UnmarshalJSON(b []byte) error {
	var raw_publication string
	err := json.Unmarshal(b, &raw_publication)
	if err != nil {
		return err
	}

	parts := strings.SplitN(raw_publication, ":", 2)

	switch parts[0] {
	case "pubmed",
		"doi",
		"patent",
		"url":
		p.Type = parts[0]
		p.Reference = parts[1]
		return nil
	default:
		return fmt.Errorf("invalid publication type %s", parts[0])
	}
}

func (p *Publication) MarshalJSON() ([]byte, error) {
	buffer := &bytes.Buffer{}
	raw_publication := fmt.Sprintf("%s:%s", p.Type, p.Reference)
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(raw_publication)
	return buffer.Bytes(), err
}

type Cluster struct {
	Alkaloid            *Alkaloid           `json:"alkaloid,omitempty"`
	BiosyntheticClasses []BiosyntheticClass `json:"biosyn_class"`
	Compounds           []ChemCompound      `json:"compounds"`
	Genes               *Genes              `json:"genes,omitempty"`
	Loci                Loci                `json:"loci"`
	MibigAccession      string              `json:"mibig_accession"`
	Minimal             bool                `json:"minimal"`
	NcbiTaxId           string              `json:"ncbi_tax_id"`
	Nrp                 *Nrp                `json:"nrp,omitempty"`
	OrganismName        string              `json:"organism_name"`
	Other               *Other              `json:"other,omitempty"`
	Polyketide          *Polyketide         `json:"polyketide,omitempty"`
	Publications        []Publication       `json:"publications,omitempty"`
	RetirementReasons   []string            `json:"retirement_reasons,omitempty"`
	RiPP                *RiPP               `json:"ripp,omitempty"`
	Saccharide          *Saccharide         `json:"saccharide,omitempty"`
	SeeAlso             []string            `json:"see_also,omitempty"`
	Status              string              `json:"status"`
	Terpene             *Terpene            `json:"terpene,omitempty"`
}

type MibigEntry struct {
	ChangeLogs []ChangeLog `json:"changelog"`
	Cluster    Cluster     `json:"cluster"`
	Comments   string      `json:"comments,omitempty"` // TODO: Do we still need this, now that we have the changelog?
}

type MibigEntryStatus struct {
	EntryId string
	Status  string
	Reason  string
	See     []string
}
