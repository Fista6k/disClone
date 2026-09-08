package domain

import "errors"

var (
	ErrUserNotFound     = errors.New("User not found")
	ErrUserExists       = errors.New("User already exists")
	ErrIncorrectassword = errors.New("Incorrect password")
)
