/*
Copyright © 2024 Technical University of Denmark - written by Kai Blin <kblin@biosustain.dtu.dk>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"secondarymetabolites.org/mibig-api/internal/models"
)

// repoListCmd represents the repoList command
var repoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all entries in the repository",
	Long: `List all entries in the repository.

This will list all entries currently in the repository, including pending and embargoed entries.`,
	Run: func(cmd *cobra.Command, args []string) {
		listEntries()
	},
}

func listEntries() {
	db, err := InitDb()
	if err != nil {
		panic(fmt.Errorf("error opening database: %s", err))
	}

	m := models.NewModels(db)

	entries, err := m.Entries.List()
	if err != nil {
		panic(fmt.Errorf("error reading entries: %s", err))
	}

	for _, entry := range entries {
		fmt.Printf("%s.%d\t%s\t%s\t%s\t%d\t%s\t%s\t%s\n", entry.Accession, entry.Version, entry.Status, entry.Quality, entry.Completeness, entry.Taxonomy.NcbiTaxId, entry.Taxonomy.Name, entry.RetirementReasons, entry.SeeAlso)
	}
}

func init() {
	repoCmd.AddCommand(repoListCmd)
}
