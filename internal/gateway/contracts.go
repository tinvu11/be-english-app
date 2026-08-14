// Package gateway defines ports for external services used by application use cases.
package gateway

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
)

// YouTubeProvider exposes metadata preview and normalized audio extraction.
type YouTubeProvider interface {
	PreviewVideo(ctx context.Context, youtubeID string) (entity.YouTubeVideoPreview, error)
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

type LearningContentGenerator interface {
	GenerateQuizzes(ctx context.Context, input entity.QuizGenerationRequest) ([]entity.GeneratedQuiz, error)
	GenerateLocalizedContent(ctx context.Context, input entity.LocalizedContentGenerationRequest) (entity.GeneratedLocalizedContent, error)
}

// PronunciationAssessor evaluates scripted speech without exposing provider-specific APIs.
type PronunciationAssessor interface {
	Assess(ctx context.Context, input entity.PronunciationAssessmentInput) (entity.PronunciationAssessment, error)
}
