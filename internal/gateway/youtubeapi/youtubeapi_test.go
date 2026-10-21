package youtubeapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/stretchr/testify/require"
)

type audioStub struct{}

func (audioStub) PreviewVideo(context.Context, string) (entity.YouTubeVideoPreview, error) {
	panic("metadata must come from YouTube Data API")
}

func (audioStub) DownloadAudio(context.Context, string) ([]byte, error) {
	return []byte("audio"), nil
}

func TestPreviewVideoUsesDataAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, "/youtube/v3/videos", request.URL.Path)
		require.Equal(t, "snippet,contentDetails", request.URL.Query().Get("part"))
		require.Equal(t, "TjPFZaMe2yw", request.URL.Query().Get("id"))
		require.Equal(t, "test-key", request.URL.Query().Get("key"))
		_, _ = writer.Write([]byte(`{"items":[{"id":"TjPFZaMe2yw","snippet":{"title":"Lesson","channelId":"UC123","channelTitle":"Teacher","thumbnails":{"high":{"url":"https://example.com/thumbnail.jpg"}}},"contentDetails":{"duration":"PT1H2M3S"}}]}`))
	}))
	defer server.Close()

	provider := New(audioStub{}, "test-key", server.Client()).(*Provider)
	provider.baseURL = server.URL + "/youtube/v3/videos"
	preview, err := provider.PreviewVideo(t.Context(), "TjPFZaMe2yw")
	require.NoError(t, err)
	require.Equal(t, entity.YouTubeVideoPreview{
		YouTubeID: "TjPFZaMe2yw", Title: "Lesson", ThumbnailURL: "https://example.com/thumbnail.jpg",
		DurationSeconds: 3723, ChannelYouTubeID: "UC123", ChannelName: "Teacher",
	}, preview)

	audio, err := provider.DownloadAudio(t.Context(), "TjPFZaMe2yw")
	require.NoError(t, err)
	require.Equal(t, []byte("audio"), audio)
}

func TestPreviewVideoNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	provider := New(audioStub{}, "test-key", server.Client()).(*Provider)
	provider.baseURL = server.URL
	_, err := provider.PreviewVideo(t.Context(), "TjPFZaMe2yw")
	require.ErrorIs(t, err, entity.ErrVideoNotFound)
}
