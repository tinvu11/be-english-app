package admin

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errInvalidAdminToken = errors.New("invalid token")

type verifierStub struct{}

func (verifierStub) Verify(_ context.Context, token string) (entity.AuthIdentity, error) {
	if token != "valid-admin-token" {
		return entity.AuthIdentity{}, errInvalidAdminToken
	}

	return entity.AuthIdentity{UID: "firebase-admin", Email: "admin@example.com"}, nil
}

type usersStub struct {
	role string
}

func (stub usersStub) Authenticate(_ context.Context, identity entity.AuthIdentity) (entity.User, error) {
	return entity.User{
		ID:          "6fc48ba8-b5ca-46d9-87f1-d5e51f063763",
		FirebaseUID: identity.UID,
		Email:       identity.Email,
		Username:    "admin",
		Role:        stub.role,
	}, nil
}

func (stub usersStub) GetUser(ctx context.Context, _ string) (entity.User, error) {
	return stub.Authenticate(ctx, entity.AuthIdentity{UID: "firebase-admin", Email: "admin@example.com"})
}

func (usersStub) Register(context.Context, string, string, string) (entity.User, error) {
	return entity.User{}, entity.ErrLocalAuthDisabled
}

func (usersStub) Login(context.Context, string, string) (string, error) {
	return "", entity.ErrLocalAuthDisabled
}

type loggerStub struct{}

func (loggerStub) Debug(any, ...any)   {}
func (loggerStub) Info(string, ...any) {}
func (loggerStub) Warn(string, ...any) {}
func (loggerStub) Error(any, ...any)   {}
func (loggerStub) Fatal(any, ...any)   {}

type languagesStub struct{}

func (languagesStub) ListLanguages(context.Context) ([]entity.Language, error) {
	return []entity.Language{{ID: 1, Code: "en", Name: "English", IsActive: true}}, nil
}

func (languagesStub) CreateLanguage(_ context.Context, code, name string) (entity.Language, error) {
	return entity.Language{ID: 1, Code: code, Name: name, IsActive: true}, nil
}

func (languagesStub) UpdateLanguage(_ context.Context, id int, name string, isActive bool) (entity.Language, error) {
	return entity.Language{ID: id, Code: "en", Name: name, IsActive: isActive}, nil
}

func adminTestApp(role string) *fiber.App {
	app := fiber.New()
	NewRoutes(app.Group("/v1"), usersStub{role: role}, languagesStub{}, verifierStub{}, loggerStub{})

	return app
}

func performAdminRequest(t *testing.T, app *fiber.App, method, path, token string, body []byte) *http.Response {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := app.Test(req)
	require.NoError(t, err)

	return resp
}

func TestAdminLogin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		role   string
		body   string
		status int
	}{
		{name: "admin", role: entity.RoleAdmin, body: `{"idToken":"valid-admin-token"}`, status: http.StatusOK},
		{name: "non admin", role: entity.RoleUser, body: `{"idToken":"valid-admin-token"}`, status: http.StatusForbidden},
		{name: "invalid token", role: entity.RoleAdmin, body: `{"idToken":"invalid"}`, status: http.StatusUnauthorized},
		{name: "missing token", role: entity.RoleAdmin, body: `{}`, status: http.StatusBadRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			resp := performAdminRequest(t, adminTestApp(test.role), http.MethodPost, "/v1/admin/auth/login", "", []byte(test.body))
			defer resp.Body.Close()

			assert.Equal(t, test.status, resp.StatusCode)
		})
	}
}

func TestAdminMe(t *testing.T) {
	t.Parallel()

	t.Run("admin", func(t *testing.T) {
		t.Parallel()

		resp := performAdminRequest(t, adminTestApp(entity.RoleAdmin), http.MethodGet, "/v1/admin/me", "valid-admin-token", nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), `"role":"admin"`)
	})

	t.Run("non admin", func(t *testing.T) {
		t.Parallel()

		resp := performAdminRequest(t, adminTestApp(entity.RoleUser), http.MethodGet, "/v1/admin/me", "valid-admin-token", nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

func TestAdminLanguages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		status int
	}{
		{name: "list", method: http.MethodGet, path: "/v1/admin/languages", status: http.StatusOK},
		{name: "create", method: http.MethodPost, path: "/v1/admin/languages", body: `{"code":"vi","name":"Vietnamese"}`, status: http.StatusCreated},
		{name: "update", method: http.MethodPut, path: "/v1/admin/languages/1", body: `{"name":"English","isActive":false}`, status: http.StatusOK},
		{name: "invalid id", method: http.MethodPut, path: "/v1/admin/languages/invalid", body: `{"name":"English","isActive":true}`, status: http.StatusBadRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			resp := performAdminRequest(t, adminTestApp(entity.RoleAdmin), test.method, test.path, "valid-admin-token", []byte(test.body))
			defer resp.Body.Close()

			assert.Equal(t, test.status, resp.StatusCode)
		})
	}
}
