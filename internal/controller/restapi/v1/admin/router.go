package admin

import (
	"github.com/evrone/go-clean-template/internal/controller/restapi/middleware"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

// NewRoutes registers public and protected administrator endpoints.
func NewRoutes(apiV1 fiber.Router, users usecase.User, languages usecase.Language, levels usecase.Level, topics usecase.Topic, channels usecase.Channel, adminUsers usecase.AdminUser, videos usecase.Video, captions usecase.Caption, verifier middleware.TokenVerifier, log logger.Interface) {
	ctrl := newController(users, languages, levels, topics, channels, adminUsers, videos, captions, verifier, log)
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

	channelRoutes := protected.Group("/channels")
	channelRoutes.Get("/", ctrl.listChannels)
	channelRoutes.Post("/", ctrl.createChannel)
	channelRoutes.Put("/:id", ctrl.updateChannel)
	channelRoutes.Delete("/:id", ctrl.deleteChannel)

	userRoutes := protected.Group("/users")
	userRoutes.Get("/", ctrl.listUsers)
	userRoutes.Patch("/:id/status", ctrl.setUserActive)
	userRoutes.Patch("/:id/role", ctrl.setUserRole)

	videoRoutes := protected.Group("/videos")
	videoRoutes.Get("/", ctrl.listVideos)
	videoRoutes.Post("/youtube-preview", ctrl.previewYouTubeVideo)
	videoRoutes.Get("/:id", ctrl.getVideo)
	videoRoutes.Post("/", ctrl.createVideo)
	videoRoutes.Put("/:id", ctrl.updateVideo)
	videoRoutes.Patch("/:id/status", ctrl.setVideoStatus)
	videoRoutes.Delete("/:id", ctrl.deleteVideo)

	captionRoutes := protected.Group("/videos/:videoId/captions")
	captionRoutes.Get("/", ctrl.listCaptions)
	captionRoutes.Post("/import", ctrl.importCaptions)
	captionRoutes.Post("/import-youtube", ctrl.importYouTubeCaptions)
	captionRoutes.Post("/translate", ctrl.translateCaptions)
	captionRoutes.Post("/", ctrl.createCaption)
	captionRoutes.Put("/:captionId", ctrl.updateCaption)
	captionRoutes.Delete("/:captionId", ctrl.deleteCaption)
}
