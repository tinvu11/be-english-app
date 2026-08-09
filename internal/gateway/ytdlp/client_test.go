package ytdlp

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
)

func TestClientPreviewVideo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/videos/dQw4w9WgXcQ" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"youtubeId":"dQw4w9WgXcQ","title":"Video","manualSubtitleTracks":[]}`))
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	preview, err := client.PreviewVideo(t.Context(), "dQw4w9WgXcQ")
	if err != nil {
		t.Fatal(err)
	}
	if preview.Title != "Video" {
		t.Fatalf("unexpected title: %s", preview.Title)
	}
}

func TestClientDownloadManualSubtitleNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		http.Error(writer, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	client := New(server.URL, server.Client())
	_, err := client.DownloadManualSubtitle(t.Context(), "dQw4w9WgXcQ", "en")
	if !errors.Is(err, entity.ErrManualSubtitleNotFound) {
		t.Fatalf("expected manual subtitle not found, got %v", err)
	}
}
