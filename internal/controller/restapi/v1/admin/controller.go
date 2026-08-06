// Package admin implements the REST API endpoints reserved for administrators.
package admin

import (
	"github.com/evrone/go-clean-template/internal/controller/restapi/middleware"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/go-playground/validator/v10"
)

type controller struct {
	users     usecase.User
	languages usecase.Language
	verifier  middleware.TokenVerifier
	log       logger.Interface
	validate  *validator.Validate
}

func newController(users usecase.User, languages usecase.Language, verifier middleware.TokenVerifier, log logger.Interface) *controller {
	return &controller{
		users:     users,
		languages: languages,
		verifier:  verifier,
		log:       log,
		validate:  validator.New(validator.WithRequiredStructEnabled()),
	}
}
