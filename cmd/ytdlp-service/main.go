package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/gateway"
	"github.com/evrone/go-clean-template/internal/gateway/ytdlp"
)

func main() {
	port := env("YTDLP_SERVICE_PORT", "8081")
	timeoutSeconds, _ := strconv.Atoi(env("YTDLP_TIMEOUT_SECONDS", "45"))
	maxFileMB, _ := strconv.ParseInt(env("YTDLP_MAX_FILE_MB", "10"), 10, 64)
	provider := ytdlp.NewLocal(env("YTDLP_BINARY_PATH", "yt-dlp"), time.Duration(timeoutSeconds)*time.Second, maxFileMB<<20)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(writer http.ResponseWriter, _ *http.Request) { writer.WriteHeader(http.StatusOK) })
	mux.HandleFunc("GET /v1/videos/{youtubeID}", previewHandler(provider))
	mux.HandleFunc("GET /v1/videos/{youtubeID}/subtitles/{languageCode}", subtitleHandler(provider))
	server := &http.Server{Addr: ":" + port, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("yt-dlp service listening on :%s", port)
	log.Fatal(server.ListenAndServe())
}

func previewHandler(provider gateway.YouTubeSubtitleProvider) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		preview, err := provider.PreviewVideo(request.Context(), request.PathValue("youtubeID"))
		if err != nil {
			writeError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(writer).Encode(preview); err != nil {
			log.Printf("encode preview: %v", err)
		}
	}
}

func subtitleHandler(provider gateway.YouTubeSubtitleProvider) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		data, err := provider.DownloadManualSubtitle(request.Context(), request.PathValue("youtubeID"), request.PathValue("languageCode"))
		if err != nil {
			writeError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "text/vtt; charset=utf-8")
		_, _ = writer.Write(data)
	}
}

func writeError(writer http.ResponseWriter, err error) {
	status := http.StatusBadGateway
	if errors.Is(err, entity.ErrInvalidVideo) || errors.Is(err, entity.ErrInvalidCaption) {
		status = http.StatusBadRequest
	} else if errors.Is(err, entity.ErrManualSubtitleNotFound) {
		status = http.StatusNotFound
	}
	http.Error(writer, http.StatusText(status), status)
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
