// Package gateway defines ports for external services used by application use cases.
package gateway

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
)

// YouTubeSubtitleProvider exposes only creator-provided YouTube subtitles.
type YouTubeSubtitleProvider interface {
	PreviewVideo(ctx context.Context, youtubeID string) (entity.YouTubeVideoPreview, error)
	ListManualSubtitles(ctx context.Context, youtubeID string) ([]entity.YouTubeSubtitleTrack, error)
	DownloadManualSubtitle(ctx context.Context, youtubeID, languageCode string) ([]byte, error)
}
