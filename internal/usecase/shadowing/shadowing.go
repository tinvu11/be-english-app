package shadowing

import (
	"context"
	"strings"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/gateway"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/internal/usecase"
)

const (
	maxAudioBytes             = 5 << 20
	passingPronunciationScore = 70
	passingCompletenessScore  = 80
)

type UseCase struct {
	repo     repo.ShadowingRepo
	assessor gateway.PronunciationAssessor
}

func New(repository repo.ShadowingRepo, assessor gateway.PronunciationAssessor) usecase.Shadowing {
	return &UseCase{repo: repository, assessor: assessor}
}

func (u *UseCase) Assess(ctx context.Context, userID string, videoID, captionID int64, audio []byte, contentType, locale string) (entity.ShadowingAttempt, error) {
	if userID == "" || videoID <= 0 || captionID <= 0 || len(audio) == 0 || len(audio) > maxAudioBytes || !supportedContentType(contentType) {
		return entity.ShadowingAttempt{}, entity.ErrInvalidShadowingAudio
	}
	prompt, err := u.repo.GetPrompt(ctx, videoID, captionID)
	if err != nil {
		return entity.ShadowingAttempt{}, err
	}
	prompt.LanguageCode, err = resolveLocale(locale, prompt.LanguageCode)
	if err != nil {
		return entity.ShadowingAttempt{}, err
	}
	assessment, err := u.assessor.Assess(ctx, entity.PronunciationAssessmentInput{
		Audio: audio, ContentType: normalizeContentType(contentType), ReferenceText: prompt.ReferenceText, Locale: prompt.LanguageCode,
	})
	if err != nil {
		return entity.ShadowingAttempt{}, err
	}
	attempt, err := u.repo.SaveAttempt(ctx, userID, prompt, assessment)
	if err != nil {
		return entity.ShadowingAttempt{}, err
	}
	attempt.Passed = attempt.PronunciationScore >= passingPronunciationScore && attempt.CompletenessScore >= passingCompletenessScore
	return attempt, nil
}

func (u *UseCase) ListAttempts(ctx context.Context, userID string, videoID, captionID int64) (entity.ShadowingAttemptList, error) {
	if userID == "" || videoID <= 0 || captionID < 0 {
		return entity.ShadowingAttemptList{}, entity.ErrInvalidCaption
	}
	result, err := u.repo.ListAttempts(ctx, userID, videoID, captionID)
	for index := range result.Items {
		result.Items[index].Passed = result.Items[index].PronunciationScore >= passingPronunciationScore && result.Items[index].CompletenessScore >= passingCompletenessScore
	}
	return result, err
}

func supportedContentType(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(value, "audio/wav") || strings.HasPrefix(value, "audio/x-wav") || strings.HasPrefix(value, "audio/wave") || strings.HasPrefix(value, "audio/ogg")
}

func normalizeContentType(value string) string {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(value)), "audio/ogg") {
		return "audio/ogg; codecs=opus"
	}
	return "audio/wav; codecs=audio/pcm; samplerate=16000"
}

func resolveLocale(requested, languageCode string) (string, error) {
	value := strings.TrimSpace(requested)
	if value == "" {
		value = strings.TrimSpace(languageCode)
	}
	aliases := map[string]string{
		"ar": "ar-SA", "ca": "ca-ES", "da": "da-DK", "de": "de-DE", "en": "en-US", "es": "es-ES",
		"fi": "fi-FI", "fr": "fr-FR", "hi": "hi-IN", "it": "it-IT", "ja": "ja-JP", "ko": "ko-KR",
		"ms": "ms-MY", "nl": "nl-NL", "no": "nb-NO", "nb": "nb-NO", "pl": "pl-PL", "pt": "pt-BR",
		"ru": "ru-RU", "sv": "sv-SE", "ta": "ta-IN", "th": "th-TH", "vi": "vi-VN", "zh": "zh-CN",
	}
	if locale, ok := aliases[strings.ToLower(value)]; ok {
		return locale, nil
	}
	parts := strings.Split(value, "-")
	if len(parts) == 2 && len(parts[0]) == 2 && len(parts[1]) == 2 {
		return strings.ToLower(parts[0]) + "-" + strings.ToUpper(parts[1]), nil
	}
	return "", entity.ErrInvalidLanguage
}
