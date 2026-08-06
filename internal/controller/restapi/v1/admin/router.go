package admin

import (
	"github.com/evrone/go-clean-template/internal/controller/restapi/middleware"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

// NewRoutes registers public and protected administrator endpoints.
func NewRoutes(apiV1 fiber.Router, users usecase.User, verifier middleware.TokenVerifier, log logger.Interface) {
	ctrl := newController(users, verifier, log)
	admin := apiV1.Group("/admin")

	admin.Post("/auth/login", ctrl.login)

	protected := admin.Group("", middleware.Auth(verifier, users), middleware.AdminOnly())
	protected.Get("/me", ctrl.me)
	protected.Get("/dashboard", ctrl.dashboard)
}
