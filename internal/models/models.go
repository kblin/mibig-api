package models

import "database/sql"

type Models struct {
	Entries    EntryModel
	Roles      RoleModel
	Submitters SubmitterModel
	Tokens     TokenModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Entries:    NewEntryModel(db),
		Roles:      NewRoleModel(db),
		Submitters: NewSubmitterModel(db),
		Tokens:     NewTokenModel(db),
	}
}

func NewMockModes(tokenScopes []string) Models {
	return Models{
		Entries:    NewMockEntryModel(),
		Roles:      NewMockRoleModel(),
		Submitters: NewMockSubmitterModel(),
		Tokens:     NewMockTokenModel(tokenScopes),
	}
}
