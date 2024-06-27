package models

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"secondarymetabolites.org/mibig-api/internal/data"
	"secondarymetabolites.org/mibig-api/internal/utils"
)

type UserModel interface {
	Ping() error
	Insert(user *data.User, password string) error
	GetRolesById(role_ids []int64) ([]data.Role, error)
	GetRolesByName(role_names []string) ([]data.Role, error)
	Get(email string, active_only bool) (*data.User, error)
	Authenticate(email, password string) (*data.User, error)
	ChangePassword(userId string, password string) error
	Update(user *data.User, password string) error
	List() ([]data.User, error)
	Delete(email string) error
	GetForToken(tokenScope, tokenPlaintext string) (*data.User, error)
}

type LiveUserModel struct {
	DB            *sql.DB
	roleIdCache   map[int64]*data.Role
	roleNameCache map[string]*data.Role
}

func NewUserModel(DB *sql.DB) *LiveUserModel {
	return &LiveUserModel{
		DB:            DB,
		roleIdCache:   make(map[int64]*data.Role, 5),
		roleNameCache: make(map[string]*data.Role, 5),
	}
}

func (m *LiveUserModel) Ping() error {
	return m.DB.Ping()
}

func (m *LiveUserModel) Insert(user *data.User, password string) error {
	var err error
	user.PasswordHash, err = utils.GeneratePassword(password)
	if err != nil {
		return err
	}

	if user.Id == "" {
		user.Id, err = utils.GenerateUid(15)
		if err != nil {
			return err
		}
	}

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	statement := `INSERT INTO mibig_submitters.submitters
(user_id, email, name, call_name, institution, password_hash, is_public, gdpr_consent, active)
VALUES
($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err = tx.Exec(statement, user.Id, user.Email, user.Name, user.CallName,
		user.Institution, user.PasswordHash, user.Public, user.GDPRConsent, user.Active)
	if err != nil {
		tx.Rollback()
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "submitters_email_key"`:
			return data.ErrDuplicateEmail
		default:
			return err
		}
	}

	for _, role := range user.Roles {
		_, err = tx.Exec("INSERT INTO mibig_submitters.rel_submitters_roles VALUES($1, $2)", user.Id, role.Id)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (m *LiveUserModel) GetRolesById(role_ids []int64) ([]data.Role, error) {
	var roles []data.Role

	for _, id := range role_ids {
		role, ok := m.roleIdCache[id]
		if !ok {
			statement := `SELECT role_id, name, description FROM mibig_submitters.roles WHERE role_id = $1`
			row := m.DB.QueryRow(statement, id)
			role = &data.Role{}
			err := row.Scan(&role.Id, &role.Name, &role.Description)
			if err != nil {
				return nil, err
			}
			m.roleIdCache[id] = role
		}
		roles = append(roles, *role)
	}

	return roles, nil
}

func (m *LiveUserModel) GetRolesByName(role_names []string) ([]data.Role, error) {
	var roles []data.Role

	for _, name := range role_names {
		role, ok := m.roleNameCache[name]
		if !ok {
			statement := `SELECT role_id, description FROM mibig_submitters.roles WHERE name = $1`
			row := m.DB.QueryRow(statement, name)
			role = &data.Role{Name: name}
			err := row.Scan(&role.Id, &role.Description)
			if err != nil {
				return nil, err
			}
			m.roleNameCache[name] = role
		}
		roles = append(roles, *role)
	}

	return roles, nil
}

func (m *LiveUserModel) Get(email string, active_only bool) (*data.User, error) {
	var user data.User
	statement := `SELECT u.user_id, u.email, u.name, u.call_name, u.institution, u.password_hash, u.is_public, u.gdpr_consent, u.active, u.version, array_agg(role_id) AS role_ids 
FROM mibig_submitters.submitters AS u
LEFT JOIN mibig_submitters.rel_submitters_roles USING (user_id)
WHERE u.email = $1`
	if active_only {
		statement += " AND active = TRUE"
	}
	statement += ` GROUP BY user_id;`

	var (
		role_ids_or_null []sql.NullInt64
		role_ids         []int64
	)

	row := m.DB.QueryRow(statement, email)
	err := row.Scan(&user.Id, &user.Email, &user.Name, &user.CallName, &user.Institution,
		&user.PasswordHash, &user.Public, &user.GDPRConsent, &user.Active, &user.Version,
		pq.Array(&role_ids_or_null))
	if err != nil {
		return nil, err
	}

	for _, role_id_or_null := range role_ids_or_null {
		role_id_value, err := role_id_or_null.Value()
		if err != nil {
			return nil, err
		}
		role_id, ok := role_id_value.(int64)
		if !ok {
			continue
		}
		role_ids = append(role_ids, role_id)
	}

	user.Roles, err = m.GetRolesById(role_ids)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (m *LiveUserModel) Authenticate(email, password string) (*data.User, error) {

	user, err := m.Get(email, true)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrInvalidCredentials
		} else {
			return nil, err
		}
	}

	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, data.ErrInvalidCredentials
		} else {
			return nil, err
		}
	}

	return user, nil
}

func (m *LiveUserModel) ChangePassword(userId string, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	_, err = m.DB.Exec(`UPDATE mibig_submitters.submitters SET password_hash = $1 WHERE user_id = $2`, hashedPassword, userId)
	if err != nil {
		return err
	}

	return nil
}

func (m *LiveUserModel) Update(user *data.User, password string) error {
	tx, err := m.DB.Begin()
	if err != nil {
		log.Println("Error starting TX", err.Error())
		return err
	}

	if password == "" {
		row := tx.QueryRow(`SELECT password_hash FROM mibig_submitters.submitters WHERE user_id = $1`, user.Id)
		err = row.Scan(&user.PasswordHash)
		if err != nil {
			log.Println("Error getting hashed password", err.Error())
			return err
		}
	} else {
		user.PasswordHash, err = utils.GeneratePassword(password)
		if err != nil {
			return err
		}
	}

	statement := `UPDATE mibig_submitters.submitters SET
email = $1, name = $2, call_name = $3, password_hash = $4, institution = $5, is_public = $6, gdpr_consent = $7, active = $8, version = version + 1
WHERE user_id = $9 AND version = $10`

	args := []interface{}{
		user.Email,
		user.Name,
		user.CallName,
		user.PasswordHash,
		user.Institution,
		user.Public,
		user.GDPRConsent,
		user.Active,
		user.Id,
		user.Version,
	}
	_, err = tx.Exec(statement, args...)
	if err != nil {
		tx.Rollback()
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "submitters_email_key"`:
			return data.ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return data.ErrEditConflict
		default:
			log.Println("Error updating user", user.Id, err.Error())
		}
		return err
	}

	existing_roles, err := getExistingRoles(tx, user.Id)
	if err != nil {
		tx.Rollback()
		return err
	}

	wanted_roles, err := getWantedRoles(tx, user.Roles)
	if err != nil {
		tx.Rollback()
		return err
	}

	to_delete := utils.DifferenceInt(existing_roles, wanted_roles)
	to_add := utils.DifferenceInt(wanted_roles, existing_roles)

	for _, roleId := range to_delete {
		_, err = tx.Exec("DELETE FROM mibig_submitters.rel_submitters_roles WHERE user_id = $1 AND role_id = $2", user.Id, roleId)
		if err != nil {
			tx.Rollback()
			log.Println("Error deleting roles", err.Error())
			return err
		}
	}

	for _, roleId := range to_add {
		_, err = tx.Exec("INSERT INTO mibig_submitters.rel_submitters_roles VALUES($1, $2)", user.Id, roleId)
		if err != nil {
			tx.Rollback()
			log.Println("Error adding roles", err.Error())
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func getExistingRoles(tx *sql.Tx, userId string) ([]int, error) {
	existing_roles := make([]int, 0, 5)
	rows, err := tx.Query("SELECT role_id FROM mibig_submitters.rel_submitters_roles WHERE user_id = $1", userId)
	if err != nil {
		tx.Rollback()
		log.Println("Error getting existing roles for user", userId, err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var roleId int
		err = rows.Scan(&roleId)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		existing_roles = append(existing_roles, roleId)
	}
	return existing_roles, nil
}

func getWantedRoles(tx *sql.Tx, roles []data.Role) ([]int, error) {
	wanted_roles := make([]int, 0, 5)

	for _, role := range roles {
		wanted_roles = append(wanted_roles, role.Id)
	}

	return wanted_roles, nil
}

func (m *LiveUserModel) List() ([]data.User, error) {
	var users []data.User
	statement := `SELECT u.user_id, u.email, u.name, u.call_name, u.institution, u.password_hash, u.is_public, u.gdpr_consent, u.active, array_agg(role_id) AS role_ids 
FROM mibig_submitters.submitters AS u
LEFT JOIN mibig_submitters.rel_submitters_roles USING (user_id)
GROUP BY user_id ORDER BY user_id`
	rows, err := m.DB.Query(statement)
	if err != nil {
		// No users is not an error in this context
		if errors.Is(err, sql.ErrNoRows) {
			return users, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var submitter data.User
		var (
			role_ids_or_null []sql.NullInt64
			role_ids         []int64
		)

		err := rows.Scan(&submitter.Id, &submitter.Email, &submitter.Name, &submitter.CallName, &submitter.Institution,
			&submitter.PasswordHash, &submitter.Public, &submitter.GDPRConsent, &submitter.Active, pq.Array(&role_ids_or_null))
		if err != nil {
			return nil, err
		}

		for _, role_id_or_null := range role_ids_or_null {
			role_id_value, err := role_id_or_null.Value()
			if err != nil {
				return nil, err
			}
			role_id, ok := role_id_value.(int64)
			if !ok {
				continue
			}
			role_ids = append(role_ids, role_id)
		}

		submitter.Roles, err = m.GetRolesById(role_ids)
		if err != nil {
			return nil, err
		}

		users = append(users, submitter)
	}

	return users, nil
}

func (m *LiveUserModel) Delete(email string) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	var userId string

	row := tx.QueryRow("SELECT user_id FROM mibig_submitters.submitters WHERE email = $1", email)
	err = row.Scan(&userId)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM mibig_submitters.rel_submitters_roles WHERE user_id = $1", userId)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM mibig_submitters.submitters WHERE user_id = $1", userId)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (m *LiveUserModel) GetForToken(tokenScope, tokenPlaintext string) (*data.User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
		SELECT u.user_id, u.email, u.name, u.call_name, u.institution, u.password_hash, u.is_public, u.gdpr_consent, u.active, u.version, array_agg(role_id) AS role_ids 
		FROM mibig_submitters.submitters AS u
		LEFT JOIN mibig_submitters.rel_submitters_roles USING (user_id)
		INNER JOIN mibig_submitters.tokens AS t USING (user_id)
		WHERE t.hash = $1 AND t.scope = $2 AND t.expiry > $3
		GROUP BY user_id
	`

	var (
		user             data.User
		role_ids_or_null []sql.NullInt64
		role_ids         []int64
	)

	args := []interface{}{tokenHash[:], tokenScope, time.Now()}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.Id,
		&user.Email,
		&user.Name,
		&user.CallName,
		&user.Institution,
		&user.PasswordHash,
		&user.Public,
		&user.GDPRConsent,
		&user.Active,
		&user.Version,
		pq.Array(&role_ids_or_null))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, data.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	for _, role_id_or_null := range role_ids_or_null {
		role_id_value, err := role_id_or_null.Value()
		if err != nil {
			return nil, err
		}
		role_id, ok := role_id_value.(int64)
		if !ok {
			continue
		}
		role_ids = append(role_ids, role_id)
	}

	user.Roles, err = m.GetRolesById(role_ids)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
