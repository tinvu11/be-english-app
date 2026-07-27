// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
)

//go:generate mockgen -source=contracts.go -destination=./mocks_usecase_test.go -package=usecase_test

type (
	// Translation -.
	Translation interface {
		Translate(ctx context.Context, userID string, t entity.Translation) (entity.Translation, error)
		History(ctx context.Context, userID string) (entity.TranslationHistory, error)
	}

	// User -.
	User interface {
		Authenticate(ctx context.Context, identity entity.AuthIdentity) (entity.User, error)
		Register(ctx context.Context, username, email, password string) (entity.User, error)
		Login(ctx context.Context, email, password string) (string, error)
		GetUser(ctx context.Context, userID string) (entity.User, error)
	}

	// Task -.
	Task interface {
		Create(ctx context.Context, userID, title, description string) (entity.Task, error)
		Get(ctx context.Context, userID, taskID string) (entity.Task, error)
		List(ctx context.Context, userID string, status *entity.TaskStatus, limit, offset int) ([]entity.Task, int, error)
		Update(ctx context.Context, userID, taskID, title, description string) (entity.Task, error)
		Transition(ctx context.Context, userID, taskID string, newStatus entity.TaskStatus) (entity.Task, error)
		Delete(ctx context.Context, userID, taskID string) error
	}

	// Content -.
	Content interface {
		CreateTopic(context.Context, entity.Topic) (entity.Topic, error)
		GetTopic(context.Context, int64) (entity.Topic, error)
		ListTopics(context.Context, int, int) ([]entity.Topic, int, error)
		UpdateTopic(context.Context, int64, entity.Topic) (entity.Topic, error)
		DeleteTopic(context.Context, int64) error
		CreateLevel(context.Context, entity.Level) (entity.Level, error)
		GetLevel(context.Context, int64) (entity.Level, error)
		ListLevels(context.Context, int, int) ([]entity.Level, int, error)
		UpdateLevel(context.Context, int64, entity.Level) (entity.Level, error)
		DeleteLevel(context.Context, int64) error
		CreateChannel(context.Context, entity.Channel) (entity.Channel, error)
		GetChannel(context.Context, int64) (entity.Channel, error)
		ListChannels(context.Context, int, int) ([]entity.Channel, int, error)
		UpdateChannel(context.Context, int64, entity.Channel) (entity.Channel, error)
		DeleteChannel(context.Context, int64) error
		CreateVideo(context.Context, entity.Video) (entity.Video, error)
		GetVideo(context.Context, int64) (entity.Video, error)
		ListVideos(context.Context, int, int) ([]entity.Video, int, error)
		UpdateVideo(context.Context, int64, entity.Video) (entity.Video, error)
		DeleteVideo(context.Context, int64) error
	}
)
