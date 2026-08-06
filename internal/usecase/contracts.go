// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
)

//go:generate mockgen -source=contracts.go -destination=./mocks_usecase_test.go -package=usecase_test

type (
	// User -.
	User interface {
		Authenticate(ctx context.Context, identity entity.AuthIdentity) (entity.User, error)
		Register(ctx context.Context, username, email, password string) (entity.User, error)
		Login(ctx context.Context, email, password string) (string, error)
		GetUser(ctx context.Context, userID string) (entity.User, error)
	}

	// Language manages the platform language catalog.
	Language interface {
		ListLanguages(ctx context.Context) ([]entity.Language, error)
		CreateLanguage(ctx context.Context, code, name string) (entity.Language, error)
		UpdateLanguage(ctx context.Context, id int, name string, isActive bool) (entity.Language, error)
	}

	// Level manages language-specific proficiency levels and translations.
	Level interface {
		ListLevels(ctx context.Context, languageID *int) ([]entity.Level, error)
		CreateLevel(ctx context.Context, level entity.Level) (entity.Level, error)
		UpdateLevel(ctx context.Context, id int, level entity.Level) (entity.Level, error)
		DeleteLevel(ctx context.Context, id int) error
	}

	// Topic manages topics and their localized names.
	Topic interface {
		ListTopics(ctx context.Context) ([]entity.Topic, error)
		CreateTopic(ctx context.Context, topic entity.Topic) (entity.Topic, error)
		UpdateTopic(ctx context.Context, id int, topic entity.Topic) (entity.Topic, error)
		SetTopicActive(ctx context.Context, id int, isActive bool) (entity.Topic, error)
		DeleteTopic(ctx context.Context, id int) error
	}

	// Channel manages the YouTube channel catalog.
	Channel interface {
		ListChannels(ctx context.Context) ([]entity.Channel, error)
		CreateChannel(ctx context.Context, channel entity.Channel) (entity.Channel, error)
		UpdateChannel(ctx context.Context, id int, channel entity.Channel) (entity.Channel, error)
		DeleteChannel(ctx context.Context, id int) error
	}
)
