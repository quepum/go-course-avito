package repository

import "errors"

var (
	ErrPhoneExists     = errors.New("profile with such phone already exists")
	ErrProfileNotFound = errors.New("profile not found")
)
