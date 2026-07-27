package entity

import "time"

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// User -.
type User struct {
	ID          string    `json:"id"         example:"550e8400-e29b-41d4-a716-446655440000"`
	FirebaseUID string    `json:"-"`
	Username    string    `json:"username"    example:"johndoe"`
	Email       string    `json:"email"       example:"john@example.com"`
	Role        string    `json:"role"        example:"user"`
	CreatedAt   time.Time `json:"created_at"  example:"2026-01-01T00:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at"  example:"2026-01-01T00:00:00Z"`
} // @name entity.User
