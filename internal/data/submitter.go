package data

var AnonymousUser = &User{CallName: "Anonymous"}

type User struct {
	Id           string `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	CallName     string `json:"call_name,omitempty"`
	Institution  string `json:"institution,omitempty"`
	PasswordHash []byte `json:"-"`
	Public       bool   `json:"public,omitempty"`
	GDPRConsent  bool   `json:"gdpr_consent,omitempty"`
	Active       bool   `json:"active,omitempty"`
	Roles        []Role `json:"-"` // TODO: Do we want this
	Version      int    `json:"-"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

type Role struct {
	Id          int
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
