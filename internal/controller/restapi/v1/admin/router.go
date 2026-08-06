package admin

import (
	"github.com/evrone/go-clean-template/internal/controller/restapi/middleware"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

// NewRoutes registers public and protected administrator endpoints.
func NewRoutes(apiV1 fiber.Router, users usecase.User, languages usecase.Language, levels usecase.Level, topics usecase.Topic, verifier middleware.TokenVerifier, log logger.Interface) {
	ctrl := newController(users, languages, levels, topics, verifier, log)
	admin := apiV1.Group("/admin")

	admin.Post("/auth/login", ctrl.login)

	protected := admin.Group("", middleware.Auth(verifier, users), middleware.AdminOnly())
	protected.Get("/me", ctrl.me)
	protected.Get("/dashboard", ctrl.dashboard)

	languageRoutes := protected.Group("/languages")
	languageRoutes.Get("/", ctrl.listLanguages)
	languageRoutes.Post("/", ctrl.createLanguage)
	languageRoutes.Put("/:id", ctrl.updateLanguage)

	levelRoutes := protected.Group("/levels")
	levelRoutes.Get("/", ctrl.listLevels)
	levelRoutes.Post("/", ctrl.createLevel)
	levelRoutes.Put("/:id", ctrl.updateLevel)
	levelRoutes.Delete("/:id", ctrl.deleteLevel)

	topicRoutes := protected.Group("/topics")
	topicRoutes.Get("/", ctrl.listTopics)
	topicRoutes.Post("/", ctrl.createTopic)
	topicRoutes.Put("/:id", ctrl.updateTopic)
	topicRoutes.Patch("/:id/status", ctrl.setTopicActive)
	topicRoutes.Delete("/:id", ctrl.deleteTopic)
}
