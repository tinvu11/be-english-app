package v1

import (
	"github.com/evrone/go-clean-template/internal/controller/restapi/middleware"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// NewRoutes -.
func NewRoutes(apiV1Group fiber.Router, t usecase.Translation, u usecase.User, c usecase.Content, verifier middleware.TokenVerifier, l logger.Interface) {
	r := &V1{t: t, u: u, c: c, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	// Protected routes
	protected := apiV1Group.Group("", middleware.Auth(verifier, u))

	userGroup := protected.Group("/user")
	{
		userGroup.Get("/profile", r.profile)
	}

	registerContentRoutes(protected, r)

	translationGroup := protected.Group("/translation")
	{
		translationGroup.Get("/history", r.history)
		translationGroup.Post("/do-translate", r.doTranslate)
	}
}
