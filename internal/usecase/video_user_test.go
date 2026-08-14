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

type subtitleProviderStub struct{ preview entity.YouTubeVideoPreview }

func (s subtitleProviderStub) PreviewVideo(context.Context, string) (entity.YouTubeVideoPreview, error) {
	return s.preview, nil
}
func (subtitleProviderStub) ListManualSubtitles(context.Context, string) ([]entity.YouTubeSubtitleTrack, error) {
	return nil, nil
}
func (subtitleProviderStub) DownloadManualSubtitle(context.Context, string, string) ([]byte, error) {
	return nil, nil
}
func (subtitleProviderStub) DownloadAudio(context.Context, string) ([]byte, error) { return nil, nil }

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

	actual, reused, err := video.New(videos, users, subtitleProviderStub{preview: preview}).AddUserYouTubeVideo(
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

	_, _, err := video.New(videos, users, subtitleProviderStub{}).AddUserYouTubeVideo(t.Context(), "user-1", "dQw4w9WgXcQ")
	assert.ErrorIs(t, err, entity.ErrTargetLanguageRequired)
}
