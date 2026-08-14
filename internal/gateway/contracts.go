// Package gateway defines ports for external services used by application use cases.
package gateway

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
)

// YouTubeSubtitleProvider exposes creator-provided and automatically generated YouTube subtitles.
type YouTubeSubtitleProvider interface {
	PreviewVideo(ctx context.Context, youtubeID string) (entity.YouTubeVideoPreview, error)
	ListManualSubtitles(ctx context.Context, youtubeID string) ([]entity.YouTubeSubtitleTrack, error)
	DownloadManualSubtitle(ctx context.Context, youtubeID, languageCode string) ([]byte, error)
	DownloadAudio(ctx context.Context, youtubeID string) ([]byte, error)
}

// AudioTranscriber converts normalized audio into timestamped captions.
type AudioTranscriber interface {
	Transcribe(ctx context.Context, audio []byte, languageHint string) (entity.AudioTranscription, error)
}

// CaptionTranslator translates caption text without exposing a provider-specific API.
type CaptionTranslator interface {
	TranslateCaptions(ctx context.Context, input entity.CaptionTranslationRequest) ([]entity.TranslatedCaption, error)
}

type VocabularyTranslator interface {
	TranslateVocabulary(ctx context.Context, input entity.VocabularyTranslationRequest) (entity.VocabularyTranslation, error)
}
