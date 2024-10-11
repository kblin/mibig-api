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

// repoRefreshCmd represents the repoRefresh command
var repoRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh all materialised views and other maintenance",
	Long:  `Refresh all materialised views and other maintenance.`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := InitDb()
		if err != nil {
			panic(fmt.Errorf("error opening database: %s", err))
		}
		m := models.NewModels(db)
		err = m.Entries.Refresh()
		if err != nil {
			panic(fmt.Errorf("error refreshing views: %s", err))
		}
		fmt.Println("Done.")
	},
}

func init() {
	repoCmd.AddCommand(repoRefreshCmd)
}
