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
//
//nolint:funlen // Keeping the declarative route table together makes auth boundaries auditable.
func NewRoutes(apiV1Group fiber.Router, u usecase.User, languages usecase.Language, levels usecase.Level, topics usecase.Topic, channels usecase.Channel, adminUsers usecase.AdminUser, videos usecase.Video, captions usecase.Caption, vocabulary usecase.Vocabulary, learning usecase.LearningContent, shadowing usecase.Shadowing, iap usecase.IAP, verifier middleware.TokenVerifier, l logger.Interface) {
	r := &V1{
		u:          u,
		languages:  languages,
		levels:     levels,
		topics:     topics,
		videos:     videos,
		captions:   captions,
		vocabulary: vocabulary,
		learning:   learning,
		shadowing:  shadowing,
		iap:        iap,
		l:          l,
		v:          validator.New(validator.WithRequiredStructEnabled()),
	}
	if iap != nil {
		webhooks := apiV1Group.Group("/iap/webhooks")
		webhooks.Post("/apple", r.appleIAPWebhook)
		webhooks.Post("/google", r.googleIAPWebhook)
	}

	// Protected routes. Public routes must be registered before this prefix
	// middleware because Fiber evaluates middleware in registration order.
	protected := apiV1Group.Group("", middleware.Auth(verifier, u))
	if iap != nil {
		iapGroup := protected.Group("/iap")
		iapGroup.Post("/verify", r.verifyIAPPurchase)
		iapGroup.Get("/status", r.getIAPStatus)
	}
	vocabularyGroup := protected.Group("/vocabulary")
	{
		vocabularyGroup.Post("/translate", r.translateVocabulary)
		vocabularyGroup.Get("/overview", r.getVocabularyOverview)
		vocabularyGroup.Post("/sets", r.createVocabularySet)
		vocabularyGroup.Get("/sets", r.listVocabularySets)
		vocabularyGroup.Put("/sets/:setId", r.updateVocabularySet)
		vocabularyGroup.Delete("/sets/:setId", r.deleteVocabularySet)
		vocabularyGroup.Post("/words", r.createUserVocabulary)
		vocabularyGroup.Get("/words", r.listUserVocabularies)
		vocabularyGroup.Get("/words/unlearned", r.listUnlearnedVocabularies)
		vocabularyGroup.Put("/words/:wordId", r.updateUserVocabulary)
		vocabularyGroup.Delete("/words/:wordId", r.deleteUserVocabulary)
	}
	homeGroup := protected.Group("/home")
	{
		homeGroup.Get("/feed", r.homeFeed)
	}
	channelGroup := protected.Group("/channels")
	{
		channelGroup.Get("/:channelId/videos", r.channelVideos)
	}
	videoGroup := protected.Group("/videos")
	{
		videoGroup.Get("/search", r.searchVideos)
		videoGroup.Post("/youtube-preview", r.previewUserYouTubeVideo)
		videoGroup.Post("/import-youtube", r.importUserYouTubeVideo)
		videoGroup.Get("/:videoId/captions", r.getOriginalVideoCaptions)
		videoGroup.Get("/:videoId/captions/translation", r.getTranslatedVideoCaptions)
		videoGroup.Get("/:videoId/learning-content", r.getVideoLearningContent)
		videoGroup.Put("/:videoId/dictation-progress/:captionId", r.completeDictation)
		videoGroup.Get("/:videoId/dictation-progress", r.listCompletedDictations)
		videoGroup.Post("/:videoId/shadowing-attempts/:captionId", r.assessShadowing)
		videoGroup.Get("/:videoId/shadowing-attempts", r.listShadowingAttempts)
	}

	userGroup := protected.Group("/user")
	{
		userGroup.Get("/videos", r.listUserImportedVideos)
		userGroup.Delete("/videos/:videoId", r.removeUserImportedVideo)
		userGroup.Get("/profile", r.profile)
		userGroup.Put("/languages", r.updateUserLanguages)
		userGroup.Get("/onboarding", r.onboarding)
		userGroup.Get("/levels", r.listUserLevels)
		userGroup.Get("/watch-history", r.watchHistory)
		userGroup.Get("/watch-history/:videoId", r.watchStatus)
		userGroup.Put("/watch-history/:videoId", r.upsertWatchHistory)
		userGroup.Delete("/watch-history/:videoId", r.removeWatchHistory)
		userGroup.Get("/watch-later", r.watchLater)
		userGroup.Put("/watch-later/:videoId", r.saveWatchLater)
		userGroup.Delete("/watch-later/:videoId", r.removeWatchLater)
	}

	admincontroller.NewRoutes(apiV1Group, u, languages, levels, topics, channels, adminUsers, videos, captions, verifier, l)
}
