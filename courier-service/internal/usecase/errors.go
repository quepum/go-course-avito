package usecase

import "errors"

var (
	ErrMissingRequiredFields = errors.New("missing required field")
	ErrPhoneExists           = errors.New("profile with such phone already exists")
)
