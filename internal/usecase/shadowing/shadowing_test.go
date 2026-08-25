package shadowing

import (
	"context"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	prompt entity.ShadowingPrompt
	saved  entity.PronunciationAssessment
}

func (f *fakeRepo) GetPrompt(context.Context, int64, int64) (entity.ShadowingPrompt, error) {
	return f.prompt, nil
}

func (f *fakeRepo) SaveAttempt(_ context.Context, _ string, prompt entity.ShadowingPrompt, assessment entity.PronunciationAssessment) (entity.ShadowingAttempt, error) {
	f.saved = assessment
	return entity.ShadowingAttempt{PronunciationScore: assessment.PronunciationScore, CompletenessScore: assessment.CompletenessScore, Locale: prompt.LanguageCode}, nil
}

func (f *fakeRepo) ListAttempts(context.Context, string, int64, int64) (entity.ShadowingAttemptList, error) {
	return entity.ShadowingAttemptList{}, nil
}

type fakeAssessor struct {
	input entity.PronunciationAssessmentInput
}

type fakeQuota struct {
	err error
}

func (f *fakeQuota) Consume(context.Context, string, entity.FeatureKey) (entity.QuotaStatus, error) {
	return entity.QuotaStatus{}, f.err
}

func (f *fakeAssessor) Assess(_ context.Context, input entity.PronunciationAssessmentInput) (entity.PronunciationAssessment, error) {
	f.input = input
	return entity.PronunciationAssessment{PronunciationScore: 70, CompletenessScore: 80}, nil
}

func TestAssessUsesVideoLocaleAndPassThreshold(t *testing.T) {
	t.Parallel()
	repository := &fakeRepo{prompt: entity.ShadowingPrompt{VideoID: 1, CaptionID: 2, ReferenceText: "Hello", LanguageCode: "en"}}
	assessor := &fakeAssessor{}
	uc := &UseCase{repo: repository, assessor: assessor, quota: &fakeQuota{}}

	result, err := uc.Assess(context.Background(), "user", 1, 2, []byte("audio"), "audio/x-wav", "")
	require.NoError(t, err)
	require.True(t, result.Passed)
	require.Equal(t, "en-US", assessor.input.Locale)
	require.Equal(t, "audio/wav; codecs=audio/pcm; samplerate=16000", assessor.input.ContentType)
}

func TestAssessRejectsUnsupportedAudio(t *testing.T) {
	t.Parallel()
	uc := &UseCase{repo: &fakeRepo{}, assessor: &fakeAssessor{}}
	_, err := uc.Assess(context.Background(), "user", 1, 2, []byte("audio"), "audio/webm", "en-US")
	require.ErrorIs(t, err, entity.ErrInvalidShadowingAudio)
}

func TestAssessStopsBeforeProviderWhenQuotaIsExhausted(t *testing.T) {
	t.Parallel()

	repository := &fakeRepo{prompt: entity.ShadowingPrompt{VideoID: 1, CaptionID: 2, ReferenceText: "Hello", LanguageCode: "en"}}
	assessor := &fakeAssessor{}
	uc := &UseCase{repo: repository, assessor: assessor, quota: &fakeQuota{err: entity.ErrQuotaExceeded}}

	_, err := uc.Assess(context.Background(), "user", 1, 2, []byte("audio"), "audio/x-wav", "")

	require.ErrorIs(t, err, entity.ErrQuotaExceeded)
	require.Empty(t, assessor.input.ReferenceText)
}
