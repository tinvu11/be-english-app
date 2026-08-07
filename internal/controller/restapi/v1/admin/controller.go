// Package admin implements the REST API endpoints reserved for administrators.
package admin

import (
	"github.com/evrone/go-clean-template/internal/controller/restapi/middleware"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/go-playground/validator/v10"
)

type controller struct {
	users      usecase.User
	languages  usecase.Language
	levels     usecase.Level
	topics     usecase.Topic
	channels   usecase.Channel
	adminUsers usecase.AdminUser
	videos     usecase.Video
	captions   usecase.Caption
	verifier   middleware.TokenVerifier
	log        logger.Interface
	validate   *validator.Validate
}

func newController(users usecase.User, languages usecase.Language, levels usecase.Level, topics usecase.Topic, channels usecase.Channel, adminUsers usecase.AdminUser, videos usecase.Video, captions usecase.Caption, verifier middleware.TokenVerifier, log logger.Interface) *controller {
	return &controller{
		users:      users,
		languages:  languages,
		levels:     levels,
		topics:     topics,
		channels:   channels,
		adminUsers: adminUsers,
		videos:     videos,
		captions:   captions,
		verifier:   verifier,
		log:        log,
		validate:   validator.New(validator.WithRequiredStructEnabled()),
	}
}
