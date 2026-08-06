// Package response contains administrator API response DTOs.
package response

import "github.com/evrone/go-clean-template/internal/entity"

type Login struct {
	Authenticated bool        `json:"authenticated" example:"true"`
	User          entity.User `json:"user"`
}
