package models

import "database/sql"

type Models struct {
	Entries    EntryModel
	Roles      RoleModel
	Submitters UserModel
	Tokens     TokenModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Entries:    NewEntryModel(db),
		Roles:      NewRoleModel(db),
		Submitters: NewUserModel(db),
		Tokens:     NewTokenModel(db),
	}
}

func NewMockModes(tokenScopes []string) Models {
	return Models{
		Entries: NewMockEntryModel(),
		Roles:   NewMockRoleModel(),
		Tokens:  NewMockTokenModel(tokenScopes),
	}
}
