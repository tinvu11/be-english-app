package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

type iapHTTPStub struct {
	appleCalls int
	status     entity.IAPStatus
}

func (s *iapHTTPStub) VerifyPurchase(context.Context, string, entity.VerifyPurchaseInput) (entity.IAPStatus, error) {
	return entity.IAPStatus{}, nil
}

func (s *iapHTTPStub) GetStatus(context.Context, string) (entity.IAPStatus, error) {
	return s.status, nil
}

func (s *iapHTTPStub) HandleAppleWebhook(context.Context, string) error {
	s.appleCalls++
	return nil
}

func (s *iapHTTPStub) HandleGoogleWebhook(context.Context, []byte) error { return nil }

type iapUserStub struct {
	usecase.User
	user entity.User
}

func (s iapUserStub) Authenticate(context.Context, entity.AuthIdentity) (entity.User, error) {
	return s.user, nil
}

type iapVerifierStub struct{}

func (iapVerifierStub) Verify(context.Context, string) (entity.AuthIdentity, error) {
	return entity.AuthIdentity{UID: "firebase-user"}, nil
}

func TestIAPStatusRemainsAvailableWithoutStoreVerification(t *testing.T) {
	t.Parallel()
	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)
	for _, test := range []struct {
		name    string
		until   *time.Time
		premium bool
	}{
		{name: "free account"},
		{name: "expired account", until: &past},
		{name: "premium account", until: &future, premium: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			users := iapUserStub{user: entity.User{ID: "user-1", IsActive: true, PremiumUntil: test.until}}
			NewRoutes(app.Group("/v1"), users, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
				nil, false, iapVerifierStub{}, logger.New("error"))
			request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/iap/status", nil)
			request.Header.Set(fiber.HeaderAuthorization, "Bearer test-token")
			response, err := app.Test(request)
			require.NoError(t, err)
			defer response.Body.Close()
			require.Equal(t, http.StatusOK, response.StatusCode)
			var status entity.IAPStatus
			require.NoError(t, json.NewDecoder(response.Body).Decode(&status))
			require.Equal(t, test.premium, status.IsPremium)
			require.Nil(t, status.Subscription)
			if test.premium {
				require.True(t, status.PremiumUntil.Equal(*test.until))
			} else {
				require.Nil(t, status.PremiumUntil)
			}
		})
	}
}

func TestIAPStatusUsesSubscriptionWhenStoreVerificationIsEnabled(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	users := iapUserStub{user: entity.User{ID: "user-1", IsActive: true}}
	stub := &iapHTTPStub{status: entity.IAPStatus{IsPremium: true}}
	NewRoutes(app.Group("/v1"), users, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		stub, false, iapVerifierStub{}, logger.New("error"))
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/iap/status", nil)
	request.Header.Set(fiber.HeaderAuthorization, "Bearer test-token")
	response, err := app.Test(request)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode)
	var status entity.IAPStatus
	require.NoError(t, json.NewDecoder(response.Body).Decode(&status))
	require.True(t, status.IsPremium)
}

func TestIAPStatusRequiresAuthenticationWithoutStoreVerification(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	NewRoutes(app.Group("/v1"), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		nil, false, nil, logger.New("error"))
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/iap/status", nil)
	response, err := app.Test(request)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusUnauthorized, response.StatusCode)
}

func TestIAPWebhookIsPublicWhileVerifyIsProtected(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	stub := &iapHTTPStub{}
	NewRoutes(app.Group("/v1"), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		stub, true, nil, logger.New("error"))

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

func TestAppleWebhookIsNotRegisteredWhenAppleIAPIsDisabled(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	stub := &iapHTTPStub{}
	NewRoutes(app.Group("/v1"), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		stub, false, nil, logger.New("error"))

	registered := false
	for _, route := range app.GetRoutes(true) {
		if route.Method == http.MethodPost && route.Path == "/v1/iap/webhooks/apple" {
			registered = true
			break
		}
	}

	require.False(t, registered)
	require.Zero(t, stub.appleCalls)
}
