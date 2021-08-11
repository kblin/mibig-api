/*
Copyright © 2021 Kai Blin <kblin@biosustain.dtu.dk>

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
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"secondarymetabolites.org/mibig-api/internal/data"
	"secondarymetabolites.org/mibig-api/internal/models"
)

var (
	status string
)

func getValidStates() []string {
	return []string{"published", "retired", "embargoed", "reserved"}
}

// repoImportCmd represents the repoImport command
var repoImportCmd = &cobra.Command{
	Use:   "import <json file>",
	Short: "Import a MIBiG JSON file into the database",
	Long: `Import a MIBiG JSON file into the database.

JSON files are assumed to validate against the JSON schema.
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		valid := false
		for _, validState := range getValidStates() {
			if status == validState {
				valid = true
				break
			}
		}

		if !valid {
			fmt.Fprintf(os.Stderr, "invalid status '%s', needs to be one of %v", status, getValidStates())
			os.Exit(2)
		}

		jsonFileName := args[0]

		jsonBytes, err := os.ReadFile(jsonFileName)
		if err != nil {
			panic(err)
		}

		var Entry data.MibigEntry

		if err = json.Unmarshal(jsonBytes, &Entry); err != nil {
			panic(err)
		}

		Status := data.MibigEntryStatus{
			EntryId: Entry.Cluster.MibigAccession,
			Status:  status,
		}
		db, err := InitDb()
		if err != nil {
			panic(fmt.Errorf("error opening database: %s", err))
		}

		m := models.NewModels(db)

		err = m.Entries.InsertEntryStatus(Status)
		if err != nil {
			panic(fmt.Errorf("error writing entry status: %s", err))
		}

		err = m.Entries.Add(Entry)
		if err != nil {
			panic(fmt.Errorf("error writing entry %s %s to database: %s", Entry.Cluster.MibigAccession, Entry.Cluster.NcbiTaxId, err))
		}

		buf := new(bytes.Buffer)
		encoder := json.NewEncoder(buf)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "    ")
		encoder.Encode(Entry)

		fmt.Printf("%s", strings.TrimSuffix(buf.String(), "\n"))

	},
}

func init() {
	repoCmd.AddCommand(repoImportCmd)
	status_help := fmt.Sprintf("Status of the entry to be loaded %v", getValidStates())
	repoImportCmd.Flags().StringVarP(&status, "status", "s", "published", status_help)
}
