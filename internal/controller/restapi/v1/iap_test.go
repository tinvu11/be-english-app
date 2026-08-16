package v1

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

type iapHTTPStub struct{ appleCalls int }

func (s *iapHTTPStub) VerifyPurchase(context.Context, string, entity.VerifyPurchaseInput) (entity.IAPStatus, error) {
	return entity.IAPStatus{}, nil
}

func (s *iapHTTPStub) GetStatus(context.Context, string) (entity.IAPStatus, error) {
	return entity.IAPStatus{}, nil
}

func (s *iapHTTPStub) HandleAppleWebhook(context.Context, string) error {
	s.appleCalls++
	return nil
}

func (s *iapHTTPStub) HandleGoogleWebhook(context.Context, []byte) error { return nil }

func TestIAPWebhookIsPublicWhileVerifyIsProtected(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	stub := &iapHTTPStub{}
	NewRoutes(app.Group("/v1"), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		stub, nil, logger.New("error"))

	webhookRequest := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/iap/webhooks/apple",
		bytes.NewBufferString(`{"signedPayload":"signed"}`))
	webhookRequest.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	response, err := app.Test(webhookRequest)
	require.NoError(t, err)

	defer response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, 1, stub.appleCalls)

	verifyRequest := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/iap/verify",
		bytes.NewBufferString(`{"platform":"ios","product_id":"premium","purchase_token":"1"}`))
	verifyRequest.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	response, err = app.Test(verifyRequest)
	require.NoError(t, err)

	defer response.Body.Close()
	require.Equal(t, http.StatusUnauthorized, response.StatusCode)
}
