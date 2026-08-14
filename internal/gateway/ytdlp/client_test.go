package ytdlp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientPreviewVideo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/videos/dQw4w9WgXcQ" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"youtubeId":"dQw4w9WgXcQ","title":"Video"}`))
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
