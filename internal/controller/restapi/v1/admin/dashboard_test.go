package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

type dashboardStub struct {
	err error
}

func (s dashboardStub) GetDashboard(_ context.Context, days int) (entity.Dashboard, error) {
	return entity.Dashboard{
		GeneratedAt:  time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC),
		Days:         days,
		Totals:       entity.DashboardTotals{Users: 1234, Videos: 42},
		Daily:        []entity.DashboardDay{{Date: "2026-10-25", NewUsers: 3}},
		RecentVideos: []entity.DashboardVideo{},
	}, s.err
}

func TestDashboard(t *testing.T) {
	for _, tc := range []struct {
		name, role, token, query string
		status, days             int
	}{
		{"default period", entity.RoleAdmin, "valid-admin-token", "", 200, 7},
		{"30 days", entity.RoleAdmin, "valid-admin-token", "?days=30", 200, 30},
		{"invalid number", entity.RoleAdmin, "valid-admin-token", "?days=abc", 400, 0},
		{"unsupported period", entity.RoleAdmin, "valid-admin-token", "?days=8", 400, 0},
		{"zero", entity.RoleAdmin, "valid-admin-token", "?days=0", 400, 0},
		{"unauthenticated", entity.RoleAdmin, "", "", 401, 0},
		{"non administrator", entity.RoleUser, "valid-admin-token", "", 403, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := adminTestApp(tc.role)
			defer app.Shutdown()
			req := httptest.NewRequest(http.MethodGet, "/v1/admin/dashboard"+tc.query, nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			response, err := app.Test(req)
			require.NoError(t, err)
			defer response.Body.Close()
			require.Equal(t, tc.status, response.StatusCode)
			if tc.status == http.StatusOK {
				var data entity.Dashboard
				require.NoError(t, json.NewDecoder(response.Body).Decode(&data))
				require.Equal(t, tc.days, data.Days)
				require.Equal(t, 1234, data.Totals.Users)
				require.Equal(t, "no-store", response.Header.Get("Cache-Control"))
			}
		})
	}
}

func TestDashboardDoesNotExposeDatabaseErrors(t *testing.T) {
	app := fiber.New()
	ctrl := &controller{dashboardUseCase: dashboardStub{err: errors.New("private connection details")}, log: loggerStub{}}
	app.Get("/dashboard", ctrl.dashboard)
	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/dashboard", nil))
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusInternalServerError, response.StatusCode)
	var body map[string]string
	require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
	require.Equal(t, "could not load dashboard", body["error"])
}
