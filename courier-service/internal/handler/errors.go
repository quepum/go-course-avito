package handler

import "errors"

var (
	ErrProfileNotFound       = errors.New("profile not found")
	ErrPhoneExists           = errors.New("profile with such phone already exists")
	ErrMissingRequiredFields = errors.New("missing required field")
)
