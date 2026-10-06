package domain

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserExists        = errors.New("user already exists")
	ErrIncorrectPassword = errors.New("incorrect username or password")

	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrGroupNotFound        = errors.New("group not found")
	ErrNotGroupOwner        = errors.New("user is not the group owner")
	ErrNotGroupMember       = errors.New("user is not a group member")
	ErrUserAlreadyMember    = errors.New("user is already a group member")
)
