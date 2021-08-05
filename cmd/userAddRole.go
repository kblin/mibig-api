/*
Copyright © 2020 Kai Blin <kblin@biosustain.dtu.dk>

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

	"secondarymetabolites.org/mibig-api/internal/data"
	"secondarymetabolites.org/mibig-api/internal/models"
	"secondarymetabolites.org/mibig-api/internal/utils"
)

// userAddRoleCmd represents the userAddRole command
var userAddRoleCmd = &cobra.Command{
	Use:   "add-role <email> <role> [<role>...]",
	Short: "Add role(s) to a user",
	Long: `Add role(s) to a user.

Only roles the user doesn't have yet will be added.`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		email := args[0]
		newRoleNames := args[1:]
		db, err := InitDb()
		if err != nil {
			panic(fmt.Errorf("error opening database: %s", err))
		}

		m := models.NewModels(db)

		user, err := m.Submitters.Get(email, false)
		if err != nil {
			panic(fmt.Errorf("error reading user for %s: %s", email, err))
		}

		oldRoleNames := data.RolesToStrings(user.Roles)

		roleNames := utils.UnionString(oldRoleNames, newRoleNames)

		user.Roles, err = m.Submitters.GetRolesByName(roleNames)
		if err != nil {
			panic(fmt.Errorf("error looking up roles for %v: %s", roleNames, err))
		}

		err = m.Submitters.Update(user, "")
		if err != nil {
			panic(fmt.Errorf("error updating user: %s", err))
		}

	},
}

func init() {
	userEditCmd.AddCommand(userAddRoleCmd)
}
