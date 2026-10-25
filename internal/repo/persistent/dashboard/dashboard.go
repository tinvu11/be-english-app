// Package dashboard implements PostgreSQL administrator reporting.
package dashboard

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/pkg/postgres"
)

type Repo struct{ *postgres.Postgres }

func New(pg *postgres.Postgres) repo.DashboardRepo { return &Repo{Postgres: pg} }

// One statement keeps all metrics on the same database snapshot. Calendar days
// use Vietnam time; users.created_at is a legacy UTC timestamp without timezone.
const dashboardSQL = `WITH bounds AS (
	SELECT (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Ho_Chi_Minh')::date AS today
), dates AS (
	SELECT (today - n)::date AS day FROM bounds, generate_series($1::int - 1, 0, -1) n
), registrations AS (
	SELECT (created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Ho_Chi_Minh')::date AS day, COUNT(*) AS count
	FROM users
	WHERE created_at >= (((SELECT MIN(day) FROM dates)::timestamp AT TIME ZONE 'Asia/Ho_Chi_Minh') AT TIME ZONE 'UTC')
	GROUP BY 1
), uploads AS (
	SELECT (created_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::date AS day, COUNT(*) AS count
	FROM videos WHERE is_system AND created_at >= ((SELECT MIN(day) FROM dates)::timestamp AT TIME ZONE 'Asia/Ho_Chi_Minh')
	GROUP BY 1
), video_totals AS (
	SELECT COUNT(*) FILTER (WHERE is_system) AS total,
	COUNT(*) FILTER (WHERE NOT is_system) AS imported,
	COUNT(*) FILTER (WHERE is_system AND status='published') AS published,
	COUNT(*) FILTER (WHERE is_system AND status='draft') AS draft,
	COUNT(*) FILTER (WHERE is_system AND status='archived') AS archived
	FROM videos
)
SELECT json_build_object(
	'generatedAt', CURRENT_TIMESTAMP, 'days', $1::int,
	'totals', json_build_object(
		'users', (SELECT COUNT(*) FROM users),
		'activeUsers', (SELECT COUNT(*) FROM users WHERE is_active),
		'videos', vt.total, 'userImportedVideos', vt.imported,
		'channels', (SELECT COUNT(*) FROM channels),
		'activeChannels', (SELECT COUNT(*) FROM channels WHERE is_active),
		'captions', (SELECT COUNT(*) FROM video_captions c JOIN videos v ON v.id=c.video_id WHERE v.is_system),
		'videosWithoutCaptions', (SELECT COUNT(*) FROM videos v WHERE v.is_system AND NOT EXISTS (SELECT 1 FROM video_captions c WHERE c.video_id=v.id)),
		'languages', (SELECT COUNT(*) FROM languages WHERE is_active AND is_learnable),
		'topics', (SELECT COUNT(*) FROM topics WHERE is_active),
		'levels', (SELECT COUNT(*) FROM levels)),
	'period', json_build_object(
		'newUsers', (SELECT COALESCE(SUM(r.count),0) FROM registrations r JOIN dates d ON d.day=r.day),
		'newVideos', (SELECT COALESCE(SUM(u.count),0) FROM uploads u JOIN dates d ON d.day=u.day),
		'activeLearners', (SELECT COUNT(DISTINCT h.user_id) FROM user_watch_history h
			WHERE h.last_watched_at >= ((SELECT MIN(day) FROM dates)::timestamp AT TIME ZONE 'Asia/Ho_Chi_Minh')
			AND h.last_watched_at <= CURRENT_TIMESTAMP)),
	'videoStatus', json_build_object('published', vt.published, 'draft', vt.draft, 'archived', vt.archived),
	'daily', (SELECT json_agg(json_build_object('date', to_char(d.day,'YYYY-MM-DD'), 'newUsers', COALESCE(r.count,0), 'newVideos', COALESCE(u.count,0)) ORDER BY d.day)
		FROM dates d LEFT JOIN registrations r ON r.day=d.day LEFT JOIN uploads u ON u.day=d.day),
	'recentVideos', COALESCE((SELECT json_agg(json_build_object(
		'id', v.id, 'title', v.title, 'thumbnailUrl', COALESCE(v.thumbnail_url,''),
		'status', v.status, 'channelName', c.channel_name, 'languageCode', l.code,
		'captionCount', (SELECT COUNT(*) FROM video_captions cc WHERE cc.video_id=v.id),
		'createdAt', v.created_at) ORDER BY v.created_at DESC,v.id DESC)
		FROM (SELECT * FROM videos WHERE is_system ORDER BY created_at DESC,id DESC LIMIT 5) v
		JOIN channels c ON c.id=v.channel_id JOIN languages l ON l.id=v.language_id), '[]'::json)
) FROM video_totals vt`

func (r *Repo) GetDashboard(ctx context.Context, days int) (entity.Dashboard, error) {
	var raw []byte
	if err := r.Pool.QueryRow(ctx, dashboardSQL, days).Scan(&raw); err != nil {
		return entity.Dashboard{}, fmt.Errorf("DashboardRepo - GetDashboard: %w", err)
	}
	var result entity.Dashboard
	if err := json.Unmarshal(raw, &result); err != nil {
		return entity.Dashboard{}, fmt.Errorf("DashboardRepo - GetDashboard - decode: %w", err)
	}
	return result, nil
}
