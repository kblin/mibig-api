package data

import "errors"

var (
	ErrInvalidCategory    = errors.New("invalid search category")
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	ErrDuplicateEmail     = errors.New("models: duplicate email address")
	ErrNoCredentails      = errors.New("no credentials found")
	ErrNotImplemented     = errors.New("not implemented")
	ErrRecordNotFound     = errors.New("record not found")
	ErrEditConflict       = errors.New("edit condflict, please try again")
)
