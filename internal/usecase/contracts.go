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
		UpdateLanguages(ctx context.Context, userID string, nativeLanguageID, targetLanguageID int) (entity.User, error)
		ListWatchHistory(ctx context.Context, userID string, limit, offset int) (entity.UserVideoList, error)
		ListWatchLater(ctx context.Context, userID string, limit, offset int) (entity.UserVideoList, error)
		UpsertWatchHistory(ctx context.Context, userID string, videoID int64, lastPositionSeconds int) error
		RemoveWatchHistory(ctx context.Context, userID string, videoID int64) error
		SaveWatchLater(ctx context.Context, userID string, videoID int64) error
		RemoveWatchLater(ctx context.Context, userID string, videoID int64) error
	}

	// Language manages the platform language catalog.
	Language interface {
		ListLanguages(ctx context.Context) ([]entity.Language, error)
		CreateLanguage(ctx context.Context, code, name, flagEmoji string, isLearnable bool) (entity.Language, error)
		UpdateLanguage(ctx context.Context, id int, name, flagEmoji string, isActive, isLearnable bool) (entity.Language, error)
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

	// AdminUser manages application users from the administrator console.
	AdminUser interface {
		ListUsers(ctx context.Context, filter entity.UserFilter) (entity.UserList, error)
		SetUserActive(ctx context.Context, actorID, userID string, isActive bool) (entity.User, error)
		SetUserRole(ctx context.Context, actorID, userID, role string) (entity.User, error)
	}

	// Video manages videos and their related catalog data.
	Video interface {
		ListVideos(ctx context.Context, filter entity.VideoFilter) (entity.VideoList, error)
		GetVideo(ctx context.Context, id int64) (entity.Video, error)
		PreviewYouTubeVideo(ctx context.Context, youtubeURLOrID string) (entity.YouTubeVideoPreview, error)
		AddUserYouTubeVideo(ctx context.Context, userID, youtubeURLOrID string) (entity.Video, bool, error)
		CreateVideo(ctx context.Context, input entity.VideoInput) (entity.Video, error)
		UpdateVideo(ctx context.Context, id int64, input entity.VideoInput) (entity.Video, error)
		TransitionVideoStatus(ctx context.Context, id int64, status string) (entity.Video, error)
		DeleteVideo(ctx context.Context, id int64) error
	}

	// Caption manages source captions, pronunciation guides and translations.
	Caption interface {
		ListCaptions(ctx context.Context, videoID int64, filter entity.CaptionFilter) (entity.CaptionList, error)
		CreateCaption(ctx context.Context, videoID int64, input entity.CaptionInput) (entity.Caption, error)
		ImportSRT(ctx context.Context, videoID int64, original []byte, translations []entity.SRTTranslationFile) ([]entity.Caption, error)
		ListYouTubeSubtitleTracks(ctx context.Context, videoID int64) (entity.YouTubeSubtitleTracks, error)
		ImportFromYouTube(ctx context.Context, videoID int64, languageCode, mode string) (entity.YouTubeCaptionImportResult, error)
		GetOriginalCaptions(ctx context.Context, videoID int64) (entity.VideoCaptions, error)
		GetTranslatedCaptions(ctx context.Context, userID string, videoID int64) (entity.VideoCaptions, error)
		TranslateCaptions(ctx context.Context, videoID int64, targetLanguageID int, mode string) (entity.CaptionTranslationResult, error)
		UpdateCaption(ctx context.Context, videoID, captionID int64, input entity.CaptionInput) (entity.Caption, error)
		DeleteCaption(ctx context.Context, videoID, captionID int64) error
	}

	Vocabulary interface {
		TranslateWord(ctx context.Context, userID, word string) (entity.VocabularyLookupResult, error)
	}
)
