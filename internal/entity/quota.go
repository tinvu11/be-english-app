package entity

import (
	"fmt"
	"time"
)

type FeatureKey string

const (
	FeatureYouTubeImport       FeatureKey = "youtube_import"
	FeatureShadowingAssessment FeatureKey = "shadowing_assessment"
)

func (f FeatureKey) Valid() bool {
	return f == FeatureYouTubeImport || f == FeatureShadowingAssessment
}

type QuotaStatus struct {
	Feature   FeatureKey `json:"feature"`
	Limit     int        `json:"limit"`
	Used      int        `json:"used"`
	Remaining int        `json:"remaining"`
	ResetAt   time.Time  `json:"resetAt"`
	Unlimited bool       `json:"unlimited"`
}

type QuotaExceededError struct {
	Status QuotaStatus
}

func (e *QuotaExceededError) Error() string {
	return fmt.Sprintf("%s: %s", ErrQuotaExceeded, e.Status.Feature)
}

func (e *QuotaExceededError) Unwrap() error { return ErrQuotaExceeded }
