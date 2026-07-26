package middleware_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evrone/go-clean-template/internal/controller/restapi/middleware"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errAuthTest = errors.New("auth error")

type fakeVerifier struct{}

func (fakeVerifier) Verify(_ context.Context, token string) (entity.AuthIdentity, error) {
	if token != "valid-firebase-token" {
		return entity.AuthIdentity{}, errAuthTest
	}

	return entity.AuthIdentity{UID: "firebase-uid-123", Email: "test@example.com"}, nil
}

type fakeUsers struct {
	err error
}

func (f fakeUsers) Authenticate(_ context.Context, identity entity.AuthIdentity) (entity.User, error) {
	if f.err != nil {
		return entity.User{}, f.err
	}

	return entity.User{ID: "local-user-id", FirebaseUID: identity.UID}, nil
}

func newTestApp(users fakeUsers) *fiber.App {
	app := fiber.New()
	app.Use(middleware.Auth(fakeVerifier{}, users))
	app.Get("/test", func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(string)
		if !ok {
			return c.SendStatus(http.StatusUnauthorized)
		}

		return c.SendString(userID)
	})

	return app
}

func TestAuthMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		authHeader     string
		users          fakeUsers
		expectedStatus int
		expectedBody   string
	}{
		{name: "missing header", expectedStatus: http.StatusUnauthorized},
		{name: "invalid format", authHeader: "Basic xxx", expectedStatus: http.StatusUnauthorized},
		{name: "invalid token", authHeader: "Bearer invalid", expectedStatus: http.StatusUnauthorized},
		{
			name:           "provision failure",
			authHeader:     "Bearer valid-firebase-token",
			users:          fakeUsers{err: errAuthTest},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "valid token",
			authHeader:     "Bearer valid-firebase-token",
			expectedStatus: http.StatusOK,
			expectedBody:   "local-user-id",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := newTestApp(tc.users)
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			if tc.expectedBody != "" {
				body, readErr := io.ReadAll(resp.Body)
				require.NoError(t, readErr)
				assert.Equal(t, tc.expectedBody, string(body))
			}
		})
	}
}
