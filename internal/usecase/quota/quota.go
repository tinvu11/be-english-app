// Package quota implements feature usage policy for free and premium users.
package quota

import (
	"context"
	"strings"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/internal/usecase"
)

type UserReader interface {
	GetByID(ctx context.Context, id string) (entity.User, error)
}

type Config struct {
	YouTubeImportDailyLimit       int
	ShadowingAssessmentDailyLimit int
}

type UseCase struct {
	repo     repo.QuotaRepo
	users    UserReader
	policies map[entity.FeatureKey]int
	now      func() time.Time
}

func New(repository repo.QuotaRepo, users UserReader, cfg Config) usecase.Quota {
	return &UseCase{
		repo:  repository,
		users: users,
		policies: map[entity.FeatureKey]int{
			entity.FeatureYouTubeImport:       cfg.YouTubeImportDailyLimit,
			entity.FeatureShadowingAssessment: cfg.ShadowingAssessmentDailyLimit,
		},
		now: time.Now,
	}
}

func (uc *UseCase) Consume(ctx context.Context, userID string, feature entity.FeatureKey) (entity.QuotaStatus, error) {
	limit, err := uc.resolveLimit(userID, feature)
	if err != nil {
		return entity.QuotaStatus{}, entity.ErrInvalidQuota
	}

	now := uc.now().UTC()
	windowStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	status := entity.QuotaStatus{Feature: feature, Limit: limit, ResetAt: windowStart.AddDate(0, 0, 1)}

	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return entity.QuotaStatus{}, err
	}

	if user.PremiumUntil != nil && user.PremiumUntil.After(now) {
		status.Limit = -1
		status.Remaining = -1
		status.Unlimited = true

		return status, nil
	}

	if limit == 0 {
		return status, &entity.QuotaExceededError{Status: status}
	}

	used, allowed, err := uc.repo.Consume(ctx, userID, feature, windowStart, limit)
	if err != nil {
		return entity.QuotaStatus{}, err
	}

	status.Used = used
	status.Remaining = max(limit-used, 0)

	if !allowed {
		return status, &entity.QuotaExceededError{Status: status}
	}

	return status, nil
}

func (uc *UseCase) resolveLimit(userID string, feature entity.FeatureKey) (int, error) {
	if strings.TrimSpace(userID) == "" || !feature.Valid() {
		return 0, entity.ErrInvalidQuota
	}

	limit, exists := uc.policies[feature]
	if !exists || limit < 0 {
		return 0, entity.ErrInvalidQuota
	}

	return limit, nil
}
