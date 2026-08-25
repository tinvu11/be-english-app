package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuotaErrorResponse(t *testing.T) {
	t.Parallel()

	resetAt := time.Date(2026, 8, 26, 0, 0, 0, 0, time.UTC)
	status := entity.QuotaStatus{
		Feature: entity.FeatureYouTubeImport, Limit: 3, Used: 3, Remaining: 0, ResetAt: resetAt,
	}
	app := fiber.New()
	app.Get("/", func(ctx *fiber.Ctx) error {
		return quotaErrorResponse(ctx, &entity.QuotaExceededError{Status: status})
	})

	httpResponse, err := app.Test(httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", http.NoBody))
	require.NoError(t, err)

	defer httpResponse.Body.Close()

	var body response.QuotaError
	require.NoError(t, json.NewDecoder(httpResponse.Body).Decode(&body))
	assert.Equal(t, http.StatusTooManyRequests, httpResponse.StatusCode)
	assert.Equal(t, "FEATURE_QUOTA_EXCEEDED", body.Code)
	assert.Equal(t, entity.FeatureYouTubeImport, body.Feature)
	assert.Equal(t, resetAt, body.ResetAt)
	assert.True(t, body.UpgradeRequired)
}
