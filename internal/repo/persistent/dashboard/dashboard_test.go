package dashboard

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

// This opt-in smoke test only reads an existing migrated database.
func TestGetDashboardPostgres(t *testing.T) {
	databaseURL := os.Getenv("DASHBOARD_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set DASHBOARD_TEST_DATABASE_URL to test against PostgreSQL")
	}
	pg, err := postgres.New(databaseURL)
	require.NoError(t, err)
	defer pg.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	require.True(t, pg.Pool.Ping(ctx) == nil, "configured PostgreSQL must be reachable")
	for _, days := range []int{7, 30} {
		result, queryErr := New(pg).GetDashboard(ctx, days)
		var pgError *pgconn.PgError
		if errors.As(queryErr, &pgError) {
			t.Fatalf("dashboard SQL failed: %s (SQLSTATE %s)", pgError.Message, pgError.Code)
		}
		// Do not include connection details from errors in test output.
		require.True(t, queryErr == nil, "dashboard query must succeed against migrated PostgreSQL")
		require.Equal(t, days, result.Days)
		require.Len(t, result.Daily, days)
		require.NotNil(t, result.RecentVideos)
		require.LessOrEqual(t, len(result.RecentVideos), 5)
		require.Equal(t, result.Totals.Videos, result.VideoStatus.Published+result.VideoStatus.Draft+result.VideoStatus.Archived)
		var users, videos int
		for i, day := range result.Daily {
			users += day.NewUsers
			videos += day.NewVideos
			if i > 0 {
				require.Less(t, result.Daily[i-1].Date, day.Date)
			}
		}
		require.Equal(t, users, result.Period.NewUsers)
		require.Equal(t, videos, result.Period.NewVideos)
		if os.Getenv("DASHBOARD_TEST_FIXTURE") == "1" {
			require.Equal(t, 4, result.Totals.Users)
			require.Equal(t, 3, result.Totals.ActiveUsers)
			require.Equal(t, 6, result.Totals.Videos)
			require.Equal(t, 1, result.Totals.UserImportedVideos)
			require.Equal(t, 2, result.Totals.Captions)
			require.Equal(t, 5, result.Totals.VideosWithoutCaptions)
			require.Equal(t, 2, result.Totals.Channels)
			require.Equal(t, 1, result.Totals.ActiveChannels)
			require.Equal(t, 1, result.Period.ActiveLearners)
			require.Equal(t, 5, result.Period.NewVideos)
			if days == 7 {
				require.Equal(t, 2, result.Period.NewUsers)
			} else {
				require.Equal(t, 3, result.Period.NewUsers)
			}
			require.Len(t, result.RecentVideos, 5)
			require.Equal(t, "Dashboard system 0", result.RecentVideos[0].Title)
			require.Equal(t, 2, result.RecentVideos[0].CaptionCount)
		}
	}
}
