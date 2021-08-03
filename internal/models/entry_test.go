package models

import (
	"database/sql"
	"io/ioutil"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/go-cmp/cmp"
	_ "github.com/lib/pq"

	"secondarymetabolites.org/mibig-api/internal/data"
	"secondarymetabolites.org/mibig-api/internal/queries"
)

type EntryModelTest struct {
	m        LiveEntryModel
	Teardown func()
}

func newEntryTestDB(t *testing.T) *EntryModelTest {
	mt := EntryModelTest{}

	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=secret dbname=mibig_test sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		t.Fatal(err)
	}
	migration, err := migrate.NewWithDatabaseInstance(
		"file://../../../migrations",
		"postgres", driver)
	if err != nil {
		t.Fatal(err)
	}

	err = migration.Up()
	if err != nil {
		t.Fatal(err)
	}

	script, err := ioutil.ReadFile("./testdata/testdata.sql")
	if err != nil {
		migration.Down()
		t.Fatal(err)
	}

	_, err = db.Exec(string(script))
	if err != nil {
		migration.Down()
		t.Fatal(err)
	}

	mt.m = LiveEntryModel{DB: db}

	mt.Teardown = func() {
		migration.Down()
		db.Close()
	}
	return &mt
}

func TestEntryModel(t *testing.T) {
	if testing.Short() {
		t.Skip("postgres: skipping integration test")
	}

	mt := newEntryTestDB(t)
	defer mt.Teardown()

	t.Run("Counts", mt.EntryModelCounts)
	t.Run("ClusterStats", mt.EntryModelClusterStats)
	t.Run("Repository", mt.EntryModelRepository)
	t.Run("Get", mt.EntryModelGet)
	t.Run("Search", mt.EntryModelSearch)
	t.Run("Available", mt.EntryModelAvailable)

}

func (mt *EntryModelTest) EntryModelCounts(t *testing.T) {
	counts, err := mt.m.Counts()
	if err != nil {
		t.Fatal(err)
	}

	if counts.Total != 2 {
		t.Errorf("want 2, got %d", counts.Total)
	}
}

func (mt *EntryModelTest) EntryModelClusterStats(t *testing.T) {
	expected := []data.StatCluster{
		{Type: "ripp", Description: "Ribosomally synthesized and post-translationally modified peptide", Count: 1, Class: "ripp"},
		{Type: "pks", Description: "Polyketide", Count: 1, Class: "pks"},
		{Type: "nrps", Description: "Nonribosomal peptide", Count: 1, Class: "nrps"},
	}

	stats, err := mt.m.ClusterStats()
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(expected, stats) {
		t.Errorf("ClusterStats unexpected results:\n%s", cmp.Diff(expected, stats))
	}
}

func (mt *EntryModelTest) EntryModelRepository(t *testing.T) {
	expected := []data.RepositoryEntry{
		{Accession: "BGC0000535", Complete: "complete", Minimal: false, Products: []data.Product{{Name: "nisin A", Synonyms: []string{"nisin"}}}, ProductTags: []data.ProductTag{{Name: "Lanthipeptide", Class: "ripp"}}, OrganismName: "Lactococcus lactis subsp. lactis"},
		{Accession: "BGC0001070", Complete: "complete", Minimal: false, Products: []data.Product{{Name: "kirromycin", Synonyms: []string{"mocimycin", "delvomycin"}}}, ProductTags: []data.ProductTag{
			{Name: "NRP", Class: "nrps"}, {Name: "Modular type I polyketide", Class: "pks"}, {Name: "Trans-AT type I polyketide", Class: "pks"},
		}, OrganismName: "Streptomyces collinus Tu 365"},
	}

	repo, err := mt.m.Repository()
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(expected, repo) {
		t.Errorf("Repository unexpected results:\n%s", cmp.Diff(expected, repo))
	}
}

func (mt *EntryModelTest) EntryModelGet(t *testing.T) {
	tests := []struct {
		Name           string
		Ids            []string
		ExpectedResult []data.RepositoryEntry
		ExpectedError  error
	}{
		{Name: "One", Ids: []string{"BGC0000535"}, ExpectedResult: []data.RepositoryEntry{
			{Accession: "BGC0000535", Complete: "complete", Minimal: false, Products: []data.Product{{Name: "nisin A", Synonyms: []string{"nisin"}}}, ProductTags: []data.ProductTag{{Name: "Lanthipeptide", Class: "ripp"}}, OrganismName: "Lactococcus lactis subsp. lactis"},
		}, ExpectedError: nil},
		{Name: "Two", Ids: []string{"BGC0000535", "BGC0001070"}, ExpectedResult: []data.RepositoryEntry{
			{Accession: "BGC0000535", Complete: "complete", Minimal: false, Products: []data.Product{{Name: "nisin A", Synonyms: []string{"nisin"}}}, ProductTags: []data.ProductTag{{Name: "Lanthipeptide", Class: "ripp"}}, OrganismName: "Lactococcus lactis subsp. lactis"},
			{Accession: "BGC0001070", Complete: "complete", Minimal: false, Products: []data.Product{{Name: "kirromycin", Synonyms: []string{"mocimycin", "delvomycin"}}}, ProductTags: []data.ProductTag{
				{Name: "NRP", Class: "nrps"}, {Name: "Modular type I polyketide", Class: "pks"}, {Name: "Trans-AT type I polyketide", Class: "pks"},
			}, OrganismName: "Streptomyces collinus Tu 365"},
		}, ExpectedError: nil},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			repo, err := mt.m.Get(tt.Ids)
			if err != tt.ExpectedError {
				t.Fatalf("Get(%v) unexpected error: want %s, got %s", tt.Ids, tt.ExpectedError, err)
			}

			if !cmp.Equal(tt.ExpectedResult, repo) {
				t.Errorf("Get(%v) unexpected results:\n%s", tt.Ids, cmp.Diff(tt.ExpectedResult, repo))
			}
		})
	}
}

func (mt *EntryModelTest) EntryModelSearch(t *testing.T) {
	tests := []struct {
		Name           string
		Query          queries.QueryTerm
		ExpectedResult []string
		ExpectedError  error
	}{
		{Name: "RiPP", Query: &queries.Expression{Category: "type", Term: "ripp"}, ExpectedResult: []string{"BGC0000535"}, ExpectedError: nil},
		{Name: "Operation/OR", Query: &queries.Operation{
			Operation: queries.OR,
			Left:      &queries.Expression{Category: "type", Term: "ripp"},
			Right:     &queries.Expression{Category: "type", Term: "nrps"},
		}, ExpectedResult: []string{"BGC0000535", "BGC0001070"}, ExpectedError: nil},
		{Name: "Guess Category", Query: &queries.Expression{Category: "unknown", Term: "ripp"}, ExpectedResult: []string{"BGC0000535"}, ExpectedError: nil},
		{Name: "Guess Invalid Category", Query: &queries.Expression{Category: "unknown", Term: "foobarbaz"}, ExpectedResult: nil, ExpectedError: data.ErrInvalidCategory},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			repo, err := mt.m.Search(tt.Query)
			if err != tt.ExpectedError {
				t.Fatalf("Search(%v) unexpected error: want %v, got %v", tt.Query, tt.ExpectedError, err)
			}

			if !cmp.Equal(tt.ExpectedResult, repo) {
				t.Errorf("Search(%v) unexpected results:\n%s", tt.Query, cmp.Diff(tt.ExpectedResult, repo))
			}
		})
	}
}

func (mt *EntryModelTest) EntryModelAvailable(t *testing.T) {
	tests := []struct {
		Name           string
		Category       string
		Term           string
		ExpectedResult []data.AvailableTerm
		ExpectedError  error
	}{
		{Name: "type", Category: "type", Term: "r", ExpectedResult: []data.AvailableTerm{
			{Val: "ripp", Desc: "Ribosomally synthesized and post-translationally modified peptide"},
		}, ExpectedError: nil},
		{Name: "invalid", Category: "foo", Term: "bar", ExpectedResult: nil, ExpectedError: data.ErrInvalidCategory},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			repo, err := mt.m.Available(tt.Category, tt.Term)
			if err != tt.ExpectedError {
				t.Fatalf("Available(%s, %s) unexpected error: want %v, got %v", tt.Category, tt.Term, tt.ExpectedError, err)
			}

			if !cmp.Equal(tt.ExpectedResult, repo) {
				t.Errorf("Available(%s, %s) unexpected results:\n%s", tt.Category, tt.Term, cmp.Diff(tt.ExpectedResult, repo))
			}
		})
	}
}
