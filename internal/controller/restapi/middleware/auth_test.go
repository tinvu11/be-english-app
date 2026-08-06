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
	if token != "valid-id" { // nosec G101
		return entity.AuthIdentity{}, errAuthTest
	}

	return entity.AuthIdentity{UID: "firebase-uid-123", Email: "test@example.com"}, nil
}

type fakeUsers struct {
	err    error
	role   string
	locked bool
}

func (f fakeUsers) Authenticate(_ context.Context, identity entity.AuthIdentity) (entity.User, error) {
	if f.err != nil {
		return entity.User{}, f.err
	}

	role := f.role
	if role == "" {
		role = entity.RoleUser
	}

	return entity.User{ID: "local-user-id", FirebaseUID: identity.UID, Role: role, IsActive: !f.locked}, nil
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

	app.Get("/admin", middleware.AdminOnly(), func(c *fiber.Ctx) error {
		return c.SendString("admin-ok")
	})

	return app
}

func runTestRequest(t *testing.T, app *fiber.App, path, authHeader string, expectedStatus int, expectedBody string) {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, http.NoBody)

	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(t, expectedStatus, resp.StatusCode)

	if expectedBody != "" {
		body, readErr := io.ReadAll(resp.Body)
		require.NoError(t, readErr)

		assert.Equal(t, expectedBody, string(body))
	}
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
			authHeader:     "Bearer valid-id",
			users:          fakeUsers{err: errAuthTest},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "valid token",
			authHeader:     "Bearer valid-id",
			expectedStatus: http.StatusOK,
			expectedBody:   "local-user-id",
		},
		{
			name:           "locked account",
			authHeader:     "Bearer valid-id",
			users:          fakeUsers{locked: true},
			expectedStatus: http.StatusForbidden,
			expectedBody:   `{"error":"account is locked"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := newTestApp(tc.users)

			runTestRequest(t, app, "/test", tc.authHeader, tc.expectedStatus, tc.expectedBody)
		})
	}
}

func TestAdminOnlyMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		authHeader     string
		users          fakeUsers
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "admin user - access granted",
			authHeader:     "Bearer valid-id",
			users:          fakeUsers{role: entity.RoleAdmin},
			expectedStatus: http.StatusOK,
			expectedBody:   "admin-ok",
		},
		{
			name:           "normal user - access forbidden",
			authHeader:     "Bearer valid-id",
			users:          fakeUsers{role: entity.RoleUser},
			expectedStatus: http.StatusForbidden,
			expectedBody:   `{"error":"forbidden: access denied"}`,
		},
		{
			name:           "unauthorized - missing token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := newTestApp(tc.users)

			runTestRequest(t, app, "/admin", tc.authHeader, tc.expectedStatus, tc.expectedBody)
		})
	}
}
