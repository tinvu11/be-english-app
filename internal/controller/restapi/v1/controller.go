package v1

import (
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/go-playground/validator/v10"
)

// V1 -.
type V1 struct {
	u          usecase.User
	languages  usecase.Language
	levels     usecase.Level
	topics     usecase.Topic
	videos     usecase.Video
	captions   usecase.Caption
	vocabulary usecase.Vocabulary
	learning   usecase.LearningContent
	shadowing  usecase.Shadowing
	iap        usecase.IAP
	l          logger.Interface
	v          *validator.Validate
}
