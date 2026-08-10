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
func NewRoutes(apiV1Group fiber.Router, u usecase.User, languages usecase.Language, levels usecase.Level, topics usecase.Topic, channels usecase.Channel, adminUsers usecase.AdminUser, videos usecase.Video, captions usecase.Caption, verifier middleware.TokenVerifier, l logger.Interface) {
	r := &V1{
		u:         u,
		languages: languages,
		levels:    levels,
		topics:    topics,
		videos:    videos,
		l:         l,
		v:         validator.New(validator.WithRequiredStructEnabled()),
	}
	// Protected routes
	protected := apiV1Group.Group("", middleware.Auth(verifier, u))
	homeGroup := protected.Group("/home")
	{
		homeGroup.Get("/feed", r.homeFeed)
	}
	channelGroup := protected.Group("/channels")
	{
		channelGroup.Get("/:channelId/videos", r.channelVideos)
	}

	userGroup := protected.Group("/user")
	{
		userGroup.Get("/profile", r.profile)
		userGroup.Put("/languages", r.updateUserLanguages)
		userGroup.Get("/onboarding", r.onboarding)
		userGroup.Get("/levels", r.listUserLevels)
		userGroup.Get("/watch-history", r.watchHistory)
		userGroup.Put("/watch-history/:videoId", r.upsertWatchHistory)
		userGroup.Delete("/watch-history/:videoId", r.removeWatchHistory)
		userGroup.Get("/watch-later", r.watchLater)
		userGroup.Put("/watch-later/:videoId", r.saveWatchLater)
		userGroup.Delete("/watch-later/:videoId", r.removeWatchLater)
	}

	admincontroller.NewRoutes(apiV1Group, u, languages, levels, topics, channels, adminUsers, videos, captions, verifier, l)
}
