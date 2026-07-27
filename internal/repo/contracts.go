// Package repo implements application outer layer logic. Each logic group in own file.
package repo

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
)

//go:generate mockgen -source=contracts.go -destination=../usecase/mocks_repo_test.go -package=usecase_test

type (
	// TranslationRepo -.
	TranslationRepo interface {
		Store(ctx context.Context, userID string, t entity.Translation) error
		GetHistory(ctx context.Context, userID string) ([]entity.Translation, error)
	}

	// TranslationWebAPI -.
	TranslationWebAPI interface {
		Translate(ctx context.Context, t entity.Translation) (entity.Translation, error)
	}

	// UserRepo -.
	UserRepo interface {
		Store(ctx context.Context, user *entity.User) error
		GetByID(ctx context.Context, id string) (entity.User, error)
		GetByEmail(ctx context.Context, email string) (entity.User, error)
		GetByFirebaseUID(ctx context.Context, firebaseUID string) (entity.User, error)
	}

	// TaskRepo -.
	TaskRepo interface {
		Store(ctx context.Context, task *entity.Task) error
		GetByID(ctx context.Context, userID, taskID string) (entity.Task, error)
		List(ctx context.Context, userID string, filter TaskFilter) ([]entity.Task, int, error)
		Update(ctx context.Context, task *entity.Task) error
		Delete(ctx context.Context, userID, taskID string) error
	}

	// TaskFilter -.
	TaskFilter struct {
		Status *entity.TaskStatus
		Limit  uint64
		Offset uint64
	}

	// ContentFilter -.
	ContentFilter struct {
		Limit  uint64
		Offset uint64
	}

	// ContentRepo -.
	ContentRepo interface {
		CreateTopic(context.Context, *entity.Topic) error
		GetTopic(context.Context, int64) (entity.Topic, error)
		ListTopics(context.Context, ContentFilter) ([]entity.Topic, int, error)
		UpdateTopic(context.Context, *entity.Topic) error
		DeleteTopic(context.Context, int64) error
		CreateLevel(context.Context, *entity.Level) error
		GetLevel(context.Context, int64) (entity.Level, error)
		ListLevels(context.Context, ContentFilter) ([]entity.Level, int, error)
		UpdateLevel(context.Context, *entity.Level) error
		DeleteLevel(context.Context, int64) error
		CreateChannel(context.Context, *entity.Channel) error
		GetChannel(context.Context, int64) (entity.Channel, error)
		ListChannels(context.Context, ContentFilter) ([]entity.Channel, int, error)
		UpdateChannel(context.Context, *entity.Channel) error
		DeleteChannel(context.Context, int64) error
		CreateVideo(context.Context, *entity.Video) error
		GetVideo(context.Context, int64) (entity.Video, error)
		ListVideos(context.Context, ContentFilter) ([]entity.Video, int, error)
		UpdateVideo(context.Context, *entity.Video) error
		DeleteVideo(context.Context, int64) error
	}
)
