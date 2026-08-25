package response

import (
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
)

type QuotaError struct {
	Error           string            `json:"error"`
	Code            string            `json:"code"`
	Feature         entity.FeatureKey `json:"feature"`
	Limit           int               `json:"limit"`
	Used            int               `json:"used"`
	Remaining       int               `json:"remaining"`
	ResetAt         time.Time         `json:"resetAt"`
	UpgradeRequired bool              `json:"upgradeRequired"`
}
