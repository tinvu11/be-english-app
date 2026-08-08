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
		UpdateFirebaseProfile(ctx context.Context, id, username, avatarURL string) error
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

	// ChannelRepo persists YouTube channels.
	ChannelRepo interface {
		ListChannels(ctx context.Context) ([]entity.Channel, error)
		CreateChannel(ctx context.Context, channel *entity.Channel) error
		UpdateChannel(ctx context.Context, channel *entity.Channel) error
		DeleteChannel(ctx context.Context, id int) error
	}

	// AdminUserRepo manages users from the administrator console.
	AdminUserRepo interface {
		ListUsers(ctx context.Context, filter entity.UserFilter) (entity.UserList, error)
		SetUserActive(ctx context.Context, id string, isActive bool) (entity.User, error)
		SetUserRole(ctx context.Context, id, role string) (entity.User, error)
	}

	// VideoRepo persists videos and returns hydrated relations.
	VideoRepo interface {
		ListVideos(ctx context.Context, filter entity.VideoFilter) (entity.VideoList, error)
		GetVideo(ctx context.Context, id int64) (entity.Video, error)
		CreateVideo(ctx context.Context, input entity.VideoInput) (entity.Video, error)
		UpdateVideo(ctx context.Context, id int64, input entity.VideoInput) (entity.Video, error)
		SetVideoStatus(ctx context.Context, id int64, expectedStatus, nextStatus string) (entity.Video, error)
		DeleteVideo(ctx context.Context, id int64) error
	}

	// CaptionRepo persists video captions and translations atomically.
	CaptionRepo interface {
		ListCaptions(ctx context.Context, videoID int64, filter entity.CaptionFilter) (entity.CaptionList, error)
		CreateCaption(ctx context.Context, videoID int64, input entity.CaptionInput) (entity.Caption, error)
		ImportCaptions(ctx context.Context, videoID int64, inputs []entity.CaptionInput) ([]entity.Caption, error)
		UpdateCaption(ctx context.Context, videoID, captionID int64, input entity.CaptionInput) (entity.Caption, error)
		DeleteCaption(ctx context.Context, videoID, captionID int64) error
	}
)
