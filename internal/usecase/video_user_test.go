package usecase_test

import (
	"context"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/usecase/video"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type youtubeProviderStub struct{ preview entity.YouTubeVideoPreview }

type allowQuota struct{}

func (allowQuota) Consume(context.Context, string, entity.FeatureKey) (entity.QuotaStatus, error) {
	return entity.QuotaStatus{}, nil
}

type rejectQuota struct{}

func (rejectQuota) Consume(context.Context, string, entity.FeatureKey) (entity.QuotaStatus, error) {
	status := entity.QuotaStatus{Feature: entity.FeatureYouTubeImport, Limit: 3, Used: 3}

	return status, &entity.QuotaExceededError{Status: status}
}

func (s youtubeProviderStub) PreviewVideo(context.Context, string) (entity.YouTubeVideoPreview, error) {
	return s.preview, nil
}
func (youtubeProviderStub) DownloadAudio(context.Context, string) ([]byte, error) { return nil, nil }

func TestAddUserYouTubeVideoUsesTargetLanguage(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	videos := NewMockVideoRepo(ctrl)
	users := NewMockUserRepo(ctrl)
	targetLanguageID := 7
	preview := entity.YouTubeVideoPreview{YouTubeID: "dQw4w9WgXcQ", Title: "Video"}
	expected := entity.Video{ID: 42, YouTubeID: preview.YouTubeID}

	users.EXPECT().GetByID(gomock.Any(), "user-1").Return(entity.User{TargetLanguageID: &targetLanguageID}, nil)
	videos.EXPECT().UpsertUserVideo(gomock.Any(), "user-1", preview, targetLanguageID).Return(expected, false, nil)

	actual, reused, err := video.New(videos, users, youtubeProviderStub{preview: preview}, allowQuota{}).AddUserYouTubeVideo(
		t.Context(), "user-1", "https://youtu.be/dQw4w9WgXcQ")
	require.NoError(t, err)
	assert.False(t, reused)
	assert.Equal(t, expected, actual)
}

func TestAddUserYouTubeVideoRequiresTargetLanguage(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	videos := NewMockVideoRepo(ctrl)
	users := NewMockUserRepo(ctrl)
	users.EXPECT().GetByID(gomock.Any(), "user-1").Return(entity.User{}, nil)

	_, _, err := video.New(videos, users, youtubeProviderStub{}, allowQuota{}).AddUserYouTubeVideo(t.Context(), "user-1", "dQw4w9WgXcQ")
	assert.ErrorIs(t, err, entity.ErrTargetLanguageRequired)
}

func TestAddUserYouTubeVideoStopsWhenQuotaIsExhausted(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	videos := NewMockVideoRepo(ctrl)
	users := NewMockUserRepo(ctrl)
	targetLanguageID := 7
	users.EXPECT().GetByID(gomock.Any(), "user-1").Return(entity.User{TargetLanguageID: &targetLanguageID}, nil)

	_, _, err := video.New(videos, users, youtubeProviderStub{}, rejectQuota{}).
		AddUserYouTubeVideo(t.Context(), "user-1", "dQw4w9WgXcQ")

	assert.ErrorIs(t, err, entity.ErrQuotaExceeded)
}

func TestListAndRemoveUserVideos(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	videos := NewMockVideoRepo(ctrl)
	users := NewMockUserRepo(ctrl)
	expected := entity.VideoList{Items: []entity.Video{{ID: 42}}, Total: 1}
	videos.EXPECT().ListUserVideos(gomock.Any(), "user-1", 20, 0).Return(expected, nil)
	videos.EXPECT().RemoveUserVideo(gomock.Any(), "user-1", int64(42)).Return(nil)

	uc := video.New(videos, users, youtubeProviderStub{}, allowQuota{})
	actual, err := uc.ListUserVideos(t.Context(), "user-1", 20, 0)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
	require.NoError(t, uc.RemoveUserVideo(t.Context(), "user-1", 42))
}

func TestUserVideoLibraryRejectsInvalidInput(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	uc := video.New(NewMockVideoRepo(ctrl), NewMockUserRepo(ctrl), youtubeProviderStub{}, allowQuota{})

	_, err := uc.ListUserVideos(t.Context(), "", 20, 0)
	assert.ErrorIs(t, err, entity.ErrInvalidVideo)
	assert.ErrorIs(t, uc.RemoveUserVideo(t.Context(), "user-1", 0), entity.ErrInvalidVideo)
}
