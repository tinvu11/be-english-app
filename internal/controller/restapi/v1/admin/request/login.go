// Package request contains administrator API request DTOs.
package request

type Login struct {
	IDToken string `json:"idToken" validate:"required" example:"firebase-id-token"`
}
