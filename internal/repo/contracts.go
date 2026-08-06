// Package repo implements application outer layer logic. Each logic group in own file.
package repo

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
)

//go:generate mockgen -source=contracts.go -destination=../usecase/mocks_repo_test.go -package=usecase_test

type (
	// UserRepo -.
	UserRepo interface {
		Store(ctx context.Context, user *entity.User) error
		GetByID(ctx context.Context, id string) (entity.User, error)
		GetByEmail(ctx context.Context, email string) (entity.User, error)
		GetByFirebaseUID(ctx context.Context, firebaseUID string) (entity.User, error)
	}

	// LanguageRepo persists languages.
	LanguageRepo interface {
		ListLanguages(ctx context.Context) ([]entity.Language, error)
		CreateLanguage(ctx context.Context, language *entity.Language) error
		UpdateLanguage(ctx context.Context, language *entity.Language) error
	}

	// LevelRepo persists levels and their translations.
	LevelRepo interface {
		ListLevels(ctx context.Context, languageID *int) ([]entity.Level, error)
		CreateLevel(ctx context.Context, level *entity.Level) error
		UpdateLevel(ctx context.Context, level *entity.Level) error
		DeleteLevel(ctx context.Context, id int) error
	}

	// TopicRepo persists topics and their translations.
	TopicRepo interface {
		ListTopics(ctx context.Context) ([]entity.Topic, error)
		CreateTopic(ctx context.Context, topic *entity.Topic) error
		UpdateTopic(ctx context.Context, topic *entity.Topic) error
		SetTopicActive(ctx context.Context, id int, isActive bool) (entity.Topic, error)
		DeleteTopic(ctx context.Context, id int) error
	}
)
