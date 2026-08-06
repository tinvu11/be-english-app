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
	ErrTopicNotFound      = errors.New("topic not found")
	ErrTopicExists        = errors.New("topic slug already exists")
	ErrTopicReferenced    = errors.New("topic is referenced by videos")
	ErrInvalidTopic       = errors.New("invalid topic")
	ErrChannelNotFound    = errors.New("channel not found")
	ErrChannelExists      = errors.New("YouTube channel already exists")
	ErrChannelReferenced  = errors.New("channel is referenced by videos")
	ErrInvalidChannel     = errors.New("invalid channel")
	ErrInvalidUserFilter  = errors.New("invalid user filter")
	ErrInvalidUserRole    = errors.New("invalid user role")
	ErrAdminSelfMutation  = errors.New("admin cannot change own status or role")
)
