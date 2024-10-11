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
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"secondarymetabolites.org/mibig-api/internal/models"
)

// repoDumpCmd represents the repoDump command
var repoDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Delete all entries from the repository",
	Long: `Delete all entries from the repository.

This clears out all entries and related tables, allowing for a new
import without affecting authentication or submission data.`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := InitDb()
		if err != nil {
			panic(fmt.Errorf("error opening database: %s", err))
		}
		m := models.NewModels(db)
		if !prompt() {
			fmt.Println("Aborting without dumping database")
			os.Exit(0)
		}
		err = m.Entries.Dump()
		if err != nil {
			panic(fmt.Errorf("error dumping entries: %s", err))
		}
		fmt.Println("Dumped all entires.")
	},
}

func prompt() bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Are you sure you want to dump all entries? [y/N] ")
		answer, err := reader.ReadString('\n')
		if err != nil {
			panic(fmt.Errorf("error reading response: %s", err))
		}
		answer = strings.TrimSpace(answer)
		if answer == "" {
			return false
		}
		answer = strings.ToLower(answer)
		if answer == "y" || answer == "yes" {
			return true
		}
		if answer == "n" || answer == "no" {
			return false
		}
	}
}

func init() {
	repoCmd.AddCommand(repoDumpCmd)
}
