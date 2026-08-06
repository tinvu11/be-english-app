package entity

import "time"

// Language is a language available in the learning platform.
type Language struct {
	ID        int       `json:"id" example:"1"`
	Code      string    `json:"code" example:"en"`
	Name      string    `json:"name" example:"English"`
	IsActive  bool      `json:"isActive" example:"true"`
	CreatedAt time.Time `json:"createdAt" example:"2026-08-06T00:00:00Z"`
} // @name entity.Language
