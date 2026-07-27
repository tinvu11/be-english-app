package entity

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrLocalAuthDisabled  = errors.New("local authentication is disabled")
	ErrContentNotFound    = errors.New("content not found")
	ErrContentConflict    = errors.New("content already exists")
	ErrContentReferenced  = errors.New("content is referenced")
	ErrInvalidReference   = errors.New("invalid content reference")
)
