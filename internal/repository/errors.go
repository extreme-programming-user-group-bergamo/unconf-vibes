package repository

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
	ErrDatabaseInit = errors.New("failed to initialize database")
)
