package v1

import (
	"github.com/evrone/go-clean-template/internal/controller/restapi/middleware"
	admincontroller "github.com/evrone/go-clean-template/internal/controller/restapi/v1/admin"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// NewRoutes -.
func NewRoutes(apiV1Group fiber.Router, u usecase.User, verifier middleware.TokenVerifier, l logger.Interface) {
	r := &V1{u: u, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	// Protected routes
	protected := apiV1Group.Group("", middleware.Auth(verifier, u))

	userGroup := protected.Group("/user")
	{
		userGroup.Get("/profile", r.profile)
	}

	admincontroller.NewRoutes(apiV1Group, u, verifier, l)
}
