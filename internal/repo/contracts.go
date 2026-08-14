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
		UpdateLanguages(ctx context.Context, id string, nativeLanguageID, targetLanguageID int) error
		ListWatchHistory(ctx context.Context, userID string, limit, offset int) (entity.UserVideoList, error)
		GetWatchHistory(ctx context.Context, userID string, videoID int64) (entity.UserVideo, bool, error)
		GetVideoState(ctx context.Context, userID string, videoID int64) (entity.VideoState, error)
		ListWatchLater(ctx context.Context, userID string, limit, offset int) (entity.UserVideoList, error)
		UpsertWatchHistory(ctx context.Context, userID string, videoID int64, lastPositionSeconds int) error
		RemoveWatchHistory(ctx context.Context, userID string, videoID int64) error
		SaveWatchLater(ctx context.Context, userID string, videoID int64) error
		RemoveWatchLater(ctx context.Context, userID string, videoID int64) error
		CompleteDictation(ctx context.Context, userID string, videoID, captionID int64) (entity.DictationProgress, error)
		ListCompletedDictations(ctx context.Context, userID string, videoID int64) (entity.DictationProgressList, error)
	}

	// LanguageRepo persists languages.
	LanguageRepo interface {
		ListLanguages(ctx context.Context) ([]entity.Language, error)
		CreateLanguage(ctx context.Context, language *entity.Language) error
		UpdateLanguage(ctx context.Context, language *entity.Language) error
		GetLanguage(ctx context.Context, id int) (entity.Language, error)
		GetLanguageByCode(ctx context.Context, code string) (entity.Language, error)
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
		UpsertUserVideo(ctx context.Context, userID string, preview entity.YouTubeVideoPreview, languageID int) (entity.Video, bool, error)
		CreateVideo(ctx context.Context, input entity.VideoInput) (entity.Video, error)
		UpdateVideo(ctx context.Context, id int64, input entity.VideoInput) (entity.Video, error)
		SetVideoStatus(ctx context.Context, id int64, expectedStatus, nextStatus string) (entity.Video, error)
		DeleteVideo(ctx context.Context, id int64) error
		UpdateVideoLanguage(ctx context.Context, id int64, languageID int) error
	}

	// CaptionRepo persists video captions and translations atomically.
	CaptionRepo interface {
		ListCaptions(ctx context.Context, videoID int64, filter entity.CaptionFilter) (entity.CaptionList, error)
		CreateCaption(ctx context.Context, videoID int64, input entity.CaptionInput) (entity.Caption, error)
		ImportCaptions(ctx context.Context, videoID int64, inputs []entity.CaptionInput) ([]entity.Caption, error)
		ImportCaptionsIfEmpty(ctx context.Context, videoID int64, inputs []entity.CaptionInput) ([]entity.Caption, error)
		ReplaceCaptions(ctx context.Context, videoID int64, inputs []entity.CaptionInput) ([]entity.Caption, error)
		UpdateCaption(ctx context.Context, videoID, captionID int64, input entity.CaptionInput) (entity.Caption, error)
		DeleteCaption(ctx context.Context, videoID, captionID int64) error
		ListCaptionsForTranslation(ctx context.Context, videoID int64, targetLanguageID int, missingOnly bool) ([]entity.Caption, int, error)
		UpsertTranslations(ctx context.Context, videoID int64, languageID int, items []entity.CaptionTranslationUpsert) error
	}

	DictionaryRepo interface {
		FindDictionaryEntry(ctx context.Context, word string, sourceLanguageID, targetLanguageID int) (entity.DictionaryEntry, error)
		UpsertDictionaryEntry(ctx context.Context, input entity.DictionaryInput) (entity.DictionaryEntry, bool, error)
		GetVocabularyOverview(ctx context.Context, userID string, languages entity.VocabularyLanguages) (entity.VocabularyOverview, error)
		CreateVocabularySet(ctx context.Context, userID, title, colorHex string, languages entity.VocabularyLanguages) (entity.VocabularySet, error)
		ListVocabularySets(ctx context.Context, userID string, languages entity.VocabularyLanguages) ([]entity.VocabularySet, error)
		UpdateVocabularySet(ctx context.Context, userID string, id int64, title, colorHex string, languages entity.VocabularyLanguages) (entity.VocabularySet, error)
		DeleteVocabularySet(ctx context.Context, userID string, id int64, languages entity.VocabularyLanguages) error
		CreateUserVocabulary(ctx context.Context, userID string, input entity.UserVocabularyInput, languages entity.VocabularyLanguages) (entity.UserVocabulary, error)
		ListUserVocabularies(ctx context.Context, userID string, filter entity.UserVocabularyFilter, languages entity.VocabularyLanguages) ([]entity.UserVocabulary, error)
		ListUnlearnedVocabularies(ctx context.Context, userID string, filter entity.UnlearnedVocabularyFilter, languages entity.VocabularyLanguages) ([]entity.UserVocabulary, error)
		UpdateUserVocabulary(ctx context.Context, userID string, id int64, input entity.UserVocabularyUpdate, languages entity.VocabularyLanguages) (entity.UserVocabulary, error)
		DeleteUserVocabulary(ctx context.Context, userID string, id int64, languages entity.VocabularyLanguages) error
	}

	LearningContentRepo interface {
		GetQuizzes(ctx context.Context, videoID int64) ([]entity.LearningQuiz, error)
		SaveQuizzesIfAbsent(ctx context.Context, videoID int64, quizzes []entity.GeneratedQuiz) error
		GetLocalizedLearningContent(ctx context.Context, videoID int64, languageID int) (entity.VideoSummary, []entity.DictionaryEntry, error)
		SaveLocalizedLearningContentIfAbsent(ctx context.Context, summary entity.VideoSummary, dictionaryIDs []int64) error
	}
)
