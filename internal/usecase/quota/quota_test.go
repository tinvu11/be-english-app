package quota

import (
	"context"
	"testing"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeQuotaRepo struct {
	used    int
	allowed bool
	calls   int
	window  time.Time
}

func (f *fakeQuotaRepo) Consume(_ context.Context, _ string, _ entity.FeatureKey, windowStart time.Time, _ int) (
	used int, allowed bool, err error,
) {
	f.calls++
	f.window = windowStart

	return f.used, f.allowed, nil
}

type fakeUserReader struct{ user *entity.User }

func (f *fakeUserReader) GetByID(context.Context, string) (entity.User, error) { return *f.user, nil }

func newTestUseCase(repository *fakeQuotaRepo, user *entity.User, cfg Config) *UseCase {
	return &UseCase{
		repo: repository, users: &fakeUserReader{user: user}, now: time.Now,
		policies: map[entity.FeatureKey]int{
			entity.FeatureYouTubeImport:       cfg.YouTubeImportDailyLimit,
			entity.FeatureShadowingAssessment: cfg.ShadowingAssessmentDailyLimit,
		},
	}
}

func TestConsumeFreeQuota(t *testing.T) {
	t.Parallel()

	repository := &fakeQuotaRepo{used: 2, allowed: true}
	uc := newTestUseCase(repository, &entity.User{}, Config{YouTubeImportDailyLimit: 3, ShadowingAssessmentDailyLimit: 5})
	uc.now = func() time.Time { return time.Date(2026, 8, 25, 16, 30, 0, 0, time.FixedZone("ICT", 7*60*60)) }

	status, err := uc.Consume(t.Context(), "user-1", entity.FeatureYouTubeImport)

	require.NoError(t, err)
	assert.Equal(t, 2, status.Used)
	assert.Equal(t, 1, status.Remaining)
	assert.Equal(t, time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC), repository.window)
	assert.Equal(t, time.Date(2026, 8, 26, 0, 0, 0, 0, time.UTC), status.ResetAt)
}

func TestConsumeRejectsExhaustedQuota(t *testing.T) {
	t.Parallel()

	repository := &fakeQuotaRepo{used: 3, allowed: false}
	uc := newTestUseCase(repository, &entity.User{}, Config{YouTubeImportDailyLimit: 3})

	status, err := uc.Consume(t.Context(), "user-1", entity.FeatureYouTubeImport)

	assert.ErrorIs(t, err, entity.ErrQuotaExceeded)
	assert.Equal(t, 3, status.Used)
	assert.Zero(t, status.Remaining)
}

func TestPremiumBypassesCounter(t *testing.T) {
	t.Parallel()

	repository := &fakeQuotaRepo{}
	until := time.Now().Add(time.Hour)
	uc := newTestUseCase(repository, &entity.User{PremiumUntil: &until}, Config{ShadowingAssessmentDailyLimit: 5})

	status, err := uc.Consume(t.Context(), "user-1", entity.FeatureShadowingAssessment)

	require.NoError(t, err)
	assert.True(t, status.Unlimited)
	assert.Equal(t, -1, status.Remaining)
	assert.Zero(t, repository.calls)
}
