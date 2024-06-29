package models

import (
	"database/sql"
	"io/ioutil"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/go-cmp/cmp"
	_ "github.com/lib/pq"

	"secondarymetabolites.org/mibig-api/internal/data"
)

type SubmitterModelTest struct {
	m        *LiveUserModel
	Teardown func()
}

func newSubmitterTestDB(t *testing.T) *SubmitterModelTest {
	mt := SubmitterModelTest{}

	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=secret dbname=mibig_test sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		t.Fatal(err)
	}
	migration, err := migrate.NewWithDatabaseInstance(
		"file://../../../migrations",
		"postgres", driver)
	if err != nil {
		t.Fatal(err)
	}

	err = migration.Up()
	if err != nil {
		t.Fatal(err)
	}

	script, err := ioutil.ReadFile("./testdata/testdata.sql")
	if err != nil {
		migration.Down()
		t.Fatal(err)
	}

	_, err = db.Exec(string(script))
	if err != nil {
		migration.Down()
		t.Fatal(err)
	}
	mt.m = NewUserModel(db)

	mt.Teardown = func() {
		migration.Down()
		db.Close()
	}
	return &mt
}

func TestSubmitterModel(t *testing.T) {
	if testing.Short() {
		t.Skip("postgres: skipping integration test")
	}

	mt := newSubmitterTestDB(t)
	defer mt.Teardown()

	t.Run("Ping", mt.Ping)
	t.Run("GetRolesById", mt.GetRolesById)
	t.Run("GetRolesByName", mt.GetRolesByName)
	t.Run("Insert", mt.Insert)
	t.Run("Get", mt.Get)
	t.Run("Authenticate", mt.Authenticate)
	t.Run("ChangePassword", mt.ChangePassword)
	t.Run("Update", mt.Update)
	t.Run("List", mt.List)
	t.Run("Delete", mt.Delete)

}

func (mt *SubmitterModelTest) Ping(t *testing.T) {
	err := mt.m.Ping()
	if err != nil {
		t.Fatal(err)
	}
}

func (mt *SubmitterModelTest) GetRolesById(t *testing.T) {
	expected := []data.Role{
		{Id: 1, Name: "admin", Description: "Users who can manage other users"},
		{Id: 2, Name: "curator", Description: "Users who can approve new entries"},
		{Id: 3, Name: "submitter", Description: "Users who can edit entries"},
		{Id: 4, Name: "guest", Description: "Users with read only access"},
	}

	roles, err := mt.m.GetRolesById([]int64{1, 2, 3, 4})
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(expected, roles) {
		t.Errorf("GetRolesById unexpected results:\n%s", cmp.Diff(expected, roles))
	}
}

func (mt *SubmitterModelTest) GetRolesByName(t *testing.T) {
	expected := []data.Role{
		{Id: 1, Name: "admin", Description: "Users who can manage other users"},
		{Id: 2, Name: "curator", Description: "Users who can approve new entries"},
		{Id: 3, Name: "submitter", Description: "Users who can edit entries"},
		{Id: 4, Name: "guest", Description: "Users with read only access"},
	}

	roles, err := mt.m.GetRolesByName([]string{"admin", "curator", "submitter", "guest"})
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(expected, roles) {
		t.Errorf("GetRolesByName unexpected results:\n%s", cmp.Diff(expected, roles))
	}
}

func (mt *SubmitterModelTest) Insert(t *testing.T) {
	passwd := "secret"
	roles, err := mt.m.GetRolesByName([]string{"guest"})
	if err != nil {
		t.Fatal(err)
	}

	submitter := data.User{
		Email:       "eve@example.org",
		Name:        "Eve User",
		CallName:    "Eve",
		Institution: "Testing",
		Public:      true,
		GDPRConsent: true,
		Active:      true,
		Roles:       roles,
	}

	err = mt.m.Insert(&submitter, passwd)
	if err != nil {
		t.Fatal(err)
	}

	if submitter.Id == "" {
		t.Errorf("Failed to set user ID on Insert")
	}
}

func (mt *SubmitterModelTest) Get(t *testing.T) {
	expected := &data.User{
		Id:           "AAAAAAAAAAAAAAAAAAAAAAAB",
		Email:        "alice@example.org",
		Name:         "Alice User",
		CallName:     "Alice",
		Institution:  "Testing",
		PasswordHash: []uint8{0x75, 0x6e, 0x75, 0x73, 0x65, 0x64},
		Public:       true,
		GDPRConsent:  true,
		Active:       true,
		Roles:        []data.Role{{Id: 1, Name: "admin", Description: "Users who can manage other users"}},
	}

	submitter, err := mt.m.Get("alice@example.org", false)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(expected, submitter) {
		t.Errorf("Get unexpected results:\n%s", cmp.Diff(expected, submitter))
	}
}

func (mt *SubmitterModelTest) Authenticate(t *testing.T) {
	eve, err := mt.m.Get("eve@example.org", false)
	if err != nil {
		t.Fatal(err)
	}

	expected := &data.User{
		Id:           eve.Id,
		Email:        "eve@example.org",
		Name:         "Eve User",
		CallName:     "Eve",
		Institution:  "Testing",
		PasswordHash: eve.PasswordHash,
		Public:       true,
		GDPRConsent:  true,
		Active:       true,
		Roles:        []data.Role{{Id: 4, Name: "guest", Description: "Users with read only access"}},
	}

	submitter, err := mt.m.Authenticate("eve@example.org", "secret")
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(expected, submitter) {
		t.Errorf("Authenticate unexpected results:\n%s", cmp.Diff(expected, submitter))
	}

	submitter, err = mt.m.Authenticate("eve@example.com", "secret")
	if err != data.ErrInvalidCredentials {
		t.Errorf("Expected %s but got %s on invalid email", data.ErrInvalidCredentials.Error(), err.Error())
	}
	if submitter != nil {
		t.Errorf("submitter is not nil after error return but %v", submitter)
	}

	submitter, err = mt.m.Authenticate("eve@example.org", "wrong")
	if err != data.ErrInvalidCredentials {
		t.Errorf("Expected %s but got %s on invalid password", data.ErrInvalidCredentials.Error(), err.Error())
	}
	if submitter != nil {
		t.Errorf("submitter is not nil after error return but %v", submitter)
	}

}

func (mt *SubmitterModelTest) ChangePassword(t *testing.T) {
	submitter, err := mt.m.Authenticate("eve@example.org", "secret")
	if err != nil {
		t.Fatal(err)
	}

	err = mt.m.ChangePassword(submitter.Id, "supersecret")
	if err != nil {
		t.Fatal(err)
	}

	_, err = mt.m.Authenticate("eve@example.org", "secret")
	if err != data.ErrInvalidCredentials {
		t.Errorf("Still managed to authenticate with old password after password change")
	}

	_, err = mt.m.Authenticate("eve@example.org", "supersecret")
	if err != nil {
		t.Fatal(err)
	}
}

func (mt *SubmitterModelTest) Update(t *testing.T) {
	eve, err := mt.m.Get("eve@example.org", false)
	if err != nil {
		t.Fatal(err)
	}

	if eve.Institution != "Testing" {
		t.Errorf("Unexpected Institution %s", eve.Institution)
	}

	eve.Institution = "Somewhere Else"
	mt.m.Update(eve, "")

	submitter, err := mt.m.Authenticate("eve@example.org", "supersecret")
	if err != nil {
		t.Fatal(err)
	}
	if submitter.Institution != eve.Institution {
		t.Errorf("Expected %s, got %s", eve.Institution, submitter.Institution)
	}

	eve.CallName = "Dr. Eve"
	mt.m.Update(eve, "secret")

	submitter, err = mt.m.Authenticate("eve@example.org", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if submitter.CallName != eve.CallName {
		t.Errorf("Expected %s, got %s", eve.CallName, submitter.CallName)
	}

}

func (mt *SubmitterModelTest) List(t *testing.T) {
	t.Fail()
}

func (mt *SubmitterModelTest) Delete(t *testing.T) {
	eve, err := mt.m.Get("eve@example.org", false)
	if err != nil {
		t.Fatal(err)
	}

	err = mt.m.Delete(eve.Email)
	if err != nil {
		t.Fatal(err)
	}

	_, err = mt.m.Get("eve@example.org", false)
	if err != sql.ErrNoRows {
		t.Errorf("Unexpected error getting deleted user. Expected %s, got %s", sql.ErrNoRows, err)
	}
}
