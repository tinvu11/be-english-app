package entity

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrLocalAuthDisabled  = errors.New("local authentication is disabled")
	ErrLanguageNotFound   = errors.New("language not found")
	ErrLanguageExists     = errors.New("language code already exists")
	ErrInvalidLanguage    = errors.New("invalid language")
	ErrLevelNotFound      = errors.New("level not found")
	ErrLevelExists        = errors.New("level already exists for language")
	ErrLevelReferenced    = errors.New("level is referenced by videos")
	ErrInvalidLevel       = errors.New("invalid level")
	ErrInvalidReference   = errors.New("invalid reference")
)
