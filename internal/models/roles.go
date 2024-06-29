package models

import (
	"database/sql"
	"errors"

	"secondarymetabolites.org/mibig-api/internal/data"
)

type RoleModel interface {
	Ping() error
	List() ([]data.Role, error)
	Add(name, description string) (int, error)
	UserCount(name string) (int, error)
	Delete(name string) error
}

type LiveRoleModel struct {
	DB *sql.DB
}

func NewRoleModel(db *sql.DB) *LiveRoleModel {
	return &LiveRoleModel{DB: db}
}

func (m *LiveRoleModel) Ping() error {
	return m.DB.Ping()
}

func (m *LiveRoleModel) List() ([]data.Role, error) {
	var roles []data.Role
	statement := `SELECT role_id, name, description FROM auth.roles`
	rows, err := m.DB.Query(statement)
	if err != nil {
		// No roles is not an error in this context
		if errors.Is(err, sql.ErrNoRows) {
			return roles, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var role data.Role
		err = rows.Scan(&role.Id, &role.Name, &role.Description)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (m *LiveRoleModel) Add(name, description string) (int, error) {
	statement := `INSERT INTO auth.roles (name, description) VALUES (?, ?)`
	ret, err := m.DB.Exec(statement, name, description)
	if err != nil {
		return 0, err
	}

	roleId, err := ret.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(roleId), nil
}

func (m *LiveRoleModel) UserCount(name string) (int, error) {
	var count int
	statement := `SELECT COUNT(role_id) FROM auth.rel_user_roles LEFT JOIN auth.roles USING (role_id)
	WHERE name = ?`
	row := m.DB.QueryRow(statement, name)
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (m *LiveRoleModel) Delete(name string) error {
	var roleId int

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	row := tx.QueryRow(`SELECT role_id FROM auth.roles WHERE name = ?`, name)
	err = row.Scan(&roleId)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`DELETE FROM auth.rel_user_roles WHERE role_id = ?`, roleId)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`DELETE FROM auth.roles WHERE role_id = ?`, roleId)
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

/* type RoleModel interface {
	Ping() error
	List() ([]data.Role, error)
	Add(name, description string) (int, error)
	UserCount(name string) (int, error)
	Delete(name string) error
} */

type MockRoleModel struct {
	RoleUsers map[string][]string
	Roles     []*data.Role
}

func NewMockRoleModel() *MockRoleModel {
	return &MockRoleModel{}
}

func (m *MockRoleModel) Ping() error {
	return nil
}

func (m *MockRoleModel) List() ([]data.Role, error) {
	return nil, data.ErrNotImplemented
}

func (m *MockRoleModel) Add(name, description string) (int, error) {
	return -1, data.ErrNotImplemented
}

func (m *MockRoleModel) UserCount(name string) (int, error) {
	return -1, data.ErrNotImplemented
}

func (m *MockRoleModel) Delete(name string) error {
	return nil
}
