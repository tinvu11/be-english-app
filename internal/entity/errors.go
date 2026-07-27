package entity

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrLocalAuthDisabled  = errors.New("local authentication is disabled")
	ErrTaskNotFound       = errors.New("task not found")
	ErrTaskForbidden      = errors.New("task does not belong to user")
	ErrInvalidTransition  = errors.New("invalid status transition")
	ErrContentNotFound    = errors.New("content not found")
	ErrContentConflict    = errors.New("content already exists")
	ErrContentReferenced  = errors.New("content is referenced")
	ErrInvalidReference   = errors.New("invalid content reference")
)
