package data

var AnonymousUser = &User{Info: UserInfo{CallName: "Anonymous"}}

type User struct {
	Id           int64    `json:"id"`
	Email        string   `json:"email"`
	PasswordHash []byte   `json:"-"`
	Active       bool     `json:"active,omitempty"`
	Info         UserInfo `json:"info"`
	Roles        []Role   `json:"-"` // TODO: Do we want this
	Version      int      `json:"-"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

type UserInfo struct {
	Id       int64  `json:"id"`
	Alias    string `json:"alias"`
	Name     string `json:"name"`
	CallName string `json:"call_name,omitempty"`
	Org1     string `json:"org1,omitempty"`
	Org2     string `json:"org2,omitempty"`
	Org3     string `json:"org3,omitempty"`
	Orcid    string `json:"orcid,omitempty"`
	Public   bool   `json:"public,omitempty"`
	Version  int    `json:"-"`
}

type Role struct {
	Id          int64
	Name        string
	Description string
}

func RolesToStrings(roles []Role) []string {
	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}
	return roleNames
}
