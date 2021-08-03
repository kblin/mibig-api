package models

import (
	"context"
	"database/sql"
	"time"

	"secondarymetabolites.org/mibig-api/internal/data"
)

type TokenModel interface {
	New(userId string, ttl time.Duration, scope string) (*data.Token, error)
	DeleteAllForUser(userId, scope string) error
}

type LiveTokenModel struct {
	DB *sql.DB
}

func NewTokenModel(db *sql.DB) *LiveTokenModel {
	return &LiveTokenModel{DB: db}
}

func (t *LiveTokenModel) New(userId string, ttl time.Duration, scope string) (*data.Token, error) {
	token, err := data.GenerateToken(userId, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = t.insert(token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (t *LiveTokenModel) insert(token *data.Token) error {
	query := `
		INSERT INTO mibig_submitters.tokens (hash, user_id, expiry, scope)
		VALUES ($1, $2, $3, $4)`

	args := []interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := t.DB.ExecContext(ctx, query, args...)
	return err
}

func (t *LiveTokenModel) DeleteAllForUser(userId, scope string) error {
	query := `
		DELETE FROM mibig_submitters.tokens
		WHERE user_id = $1 and scope = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := t.DB.ExecContext(ctx, query, userId, scope)
	return err
}

type MockTokenModel struct {
	Tokens map[string][]*data.Token
}

func NewMockTokenModel(scopes []string) *MockTokenModel {
	scopeMap := map[string][]*data.Token{}

	for _, scope := range scopes {
		scopeMap[scope] = []*data.Token{}
	}

	return &MockTokenModel{
		Tokens: scopeMap,
	}
}

func (t *MockTokenModel) New(userId string, ttl time.Duration, scope string) (*data.Token, error) {
	token, err := data.GenerateToken(userId, ttl, scope)
	if err != nil {
		return nil, err
	}

	t.Tokens[scope] = append(t.Tokens[scope], token)

	return token, nil
}

func (t *MockTokenModel) DeleteAllForUser(userId, scope string) error {
	var remaining []*data.Token
	for _, token := range t.Tokens[scope] {
		if token.UserID != userId {
			remaining = append(remaining, token)
		}
	}
	t.Tokens[scope] = remaining
	return nil
}
