// Package user implements the Postgres-backed User repository.
package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Repo -.
type Repo struct {
	*postgres.Postgres
}

// New returns a User repository instrumented with OpenTelemetry tracing spans.
func New(pg *postgres.Postgres) repo.UserRepo {
	return newTraced(&Repo{pg})
}

// Store -.
func (r *Repo) Store(ctx context.Context, user *entity.User) error {
	sql, args, err := r.Builder.
		Insert("users").
		Columns("id, firebase_uid, username, email, role, avatar_url, created_at, updated_at").
		Values(user.ID, user.FirebaseUID, user.Username, user.Email, user.Role, user.AvatarURL, user.CreatedAt, user.UpdatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("UserRepo - Store - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return entity.ErrUserAlreadyExists
		}

		return fmt.Errorf("UserRepo - Store - r.Pool.Exec: %w", err)
	}

	return nil
}

// UpdateFirebaseProfile synchronizes profile fields controlled by Firebase.
func (r *Repo) UpdateFirebaseProfile(ctx context.Context, id, username, avatarURL string) error {
	result, err := r.Pool.Exec(ctx, `UPDATE users SET username=$2,avatar_url=NULLIF($3,''),updated_at=CURRENT_TIMESTAMP WHERE id=$1`,
		id, username, avatarURL)
	if err != nil {
		return fmt.Errorf("UserRepo - UpdateFirebaseProfile: %w", err)
	}
	if result.RowsAffected() == 0 {
		return entity.ErrUserNotFound
	}

	return nil
}

// UpdateLanguages saves the native and target language selected during onboarding.
func (r *Repo) UpdateLanguages(ctx context.Context, id string, nativeLanguageID, targetLanguageID int) error {
	result, err := r.Pool.Exec(ctx, `UPDATE users
		SET native_language_id=$2,target_language_id=$3,updated_at=CURRENT_TIMESTAMP
		WHERE id=$1`, id, nativeLanguageID, targetLanguageID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return entity.ErrInvalidReference
		}
		return fmt.Errorf("UserRepo - UpdateLanguages: %w", err)
	}
	if result.RowsAffected() == 0 {
		return entity.ErrUserNotFound
	}
	return nil
}

func (r *Repo) ListWatchHistory(ctx context.Context, userID string, limit, offset int) (entity.UserVideoList, error) {
	return r.listUserVideos(ctx, `user_watch_history`, "history", userID, limit, offset)
}

func (r *Repo) GetWatchHistory(ctx context.Context, userID string, videoID int64) (entity.UserVideo, bool, error) {
	var item entity.UserVideo
	err := r.Pool.QueryRow(ctx, `SELECT v.id,v.title,v.youtube_id,COALESCE(v.thumbnail_url,''),v.duration_seconds,
		COALESCE(l.code,''),history.last_position_seconds,history.last_watched_at
		FROM user_watch_history history
		JOIN videos v ON v.id=history.video_id
		LEFT JOIN levels l ON l.id=v.level_id
		LEFT JOIN channels c ON c.id=v.channel_id
		WHERE history.user_id=$1 AND history.video_id=$2 AND v.status=$3
		AND (NOT v.is_system OR c.is_active=TRUE)`, userID, videoID, entity.VideoStatusPublished).
		Scan(&item.ID, &item.Title, &item.YouTubeID, &item.ThumbnailURL, &item.DurationSeconds,
			&item.LevelCode, &item.LastPositionSeconds, &item.LastWatchedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.UserVideo{}, false, nil
	}
	if err != nil {
		return entity.UserVideo{}, false, fmt.Errorf("UserRepo - GetWatchHistory: %w", err)
	}
	return item, true, nil
}

func (r *Repo) ListWatchLater(ctx context.Context, userID string, limit, offset int) (entity.UserVideoList, error) {
	return r.listUserVideos(ctx, `user_watch_later`, "saved", userID, limit, offset)
}

func (r *Repo) listUserVideos(ctx context.Context, table, listType, userID string, limit, offset int) (entity.UserVideoList, error) {
	positionColumn := "0"
	orderColumn := "uv.created_at"
	timestampColumn := "uv.created_at"
	if listType == "history" {
		positionColumn = "uv.last_position_seconds"
		orderColumn = "uv.last_watched_at"
		timestampColumn = "uv.last_watched_at"
	}
	query := fmt.Sprintf(`SELECT v.id,v.title,v.youtube_id,COALESCE(v.thumbnail_url,''),v.duration_seconds,
		l.code,%s,%s,COUNT(*) OVER()
		FROM %s uv
		JOIN videos v ON v.id=uv.video_id
		JOIN levels l ON l.id=v.level_id
		JOIN channels c ON c.id=v.channel_id
		WHERE uv.user_id=$1 AND v.is_system=TRUE AND v.status=$2 AND c.is_active=TRUE
		ORDER BY %s DESC,uv.id DESC LIMIT $3 OFFSET $4`, positionColumn, timestampColumn, table, orderColumn)
	rows, err := r.Pool.Query(ctx, query, userID, entity.VideoStatusPublished, limit, offset)
	if err != nil {
		return entity.UserVideoList{}, fmt.Errorf("UserRepo - listUserVideos: %w", err)
	}
	defer rows.Close()
	result := entity.UserVideoList{Items: make([]entity.UserVideo, 0)}
	for rows.Next() {
		var item entity.UserVideo
		var activityAt time.Time
		if err = rows.Scan(&item.ID, &item.Title, &item.YouTubeID, &item.ThumbnailURL, &item.DurationSeconds,
			&item.LevelCode, &item.LastPositionSeconds, &activityAt, &result.Total); err != nil {
			return entity.UserVideoList{}, fmt.Errorf("UserRepo - listUserVideos - scan: %w", err)
		}
		if listType == "history" {
			item.LastWatchedAt = activityAt
		} else {
			item.SavedAt = activityAt
		}
		result.Items = append(result.Items, item)
	}
	if err = rows.Err(); err != nil {
		return entity.UserVideoList{}, fmt.Errorf("UserRepo - listUserVideos - rows: %w", err)
	}
	return result, nil
}

func (r *Repo) UpsertWatchHistory(ctx context.Context, userID string, videoID int64, lastPositionSeconds int) error {
	result, err := r.Pool.Exec(ctx, `INSERT INTO user_watch_history(user_id,video_id,last_position_seconds,last_watched_at)
		SELECT $1,v.id,$3,CURRENT_TIMESTAMP FROM videos v
		JOIN channels c ON c.id=v.channel_id
		WHERE v.id=$2 AND v.is_system=TRUE AND v.status=$4 AND c.is_active=TRUE
		ON CONFLICT (user_id,video_id) DO UPDATE
		SET last_position_seconds=EXCLUDED.last_position_seconds,last_watched_at=CURRENT_TIMESTAMP`,
		userID, videoID, lastPositionSeconds, entity.VideoStatusPublished)
	if err != nil {
		return fmt.Errorf("UserRepo - UpsertWatchHistory: %w", err)
	}
	if result.RowsAffected() == 0 {
		return entity.ErrVideoNotFound
	}
	return nil
}

func (r *Repo) RemoveWatchHistory(ctx context.Context, userID string, videoID int64) error {
	_, err := r.Pool.Exec(ctx, `DELETE FROM user_watch_history WHERE user_id=$1 AND video_id=$2`, userID, videoID)
	if err != nil {
		return fmt.Errorf("UserRepo - RemoveWatchHistory: %w", err)
	}
	return nil
}

func (r *Repo) SaveWatchLater(ctx context.Context, userID string, videoID int64) error {
	result, err := r.Pool.Exec(ctx, `INSERT INTO user_watch_later(user_id,video_id,created_at)
		SELECT $1,v.id,CURRENT_TIMESTAMP FROM videos v
		JOIN channels c ON c.id=v.channel_id
		WHERE v.id=$2 AND v.is_system=TRUE AND v.status=$3 AND c.is_active=TRUE
		ON CONFLICT (user_id,video_id) DO UPDATE SET created_at=user_watch_later.created_at`,
		userID, videoID, entity.VideoStatusPublished)
	if err != nil {
		return fmt.Errorf("UserRepo - SaveWatchLater: %w", err)
	}
	if result.RowsAffected() == 0 {
		return entity.ErrVideoNotFound
	}
	return nil
}

func (r *Repo) RemoveWatchLater(ctx context.Context, userID string, videoID int64) error {
	_, err := r.Pool.Exec(ctx, `DELETE FROM user_watch_later WHERE user_id=$1 AND video_id=$2`, userID, videoID)
	if err != nil {
		return fmt.Errorf("UserRepo - RemoveWatchLater: %w", err)
	}
	return nil
}

func (r *Repo) CompleteDictation(ctx context.Context, userID string, videoID, captionID int64) (entity.DictationProgress, error) {
	var item entity.DictationProgress
	err := r.Pool.QueryRow(ctx, `WITH saved AS (
		INSERT INTO dictation_progress(user_id,video_id,caption_id,is_completed,completed_at)
		SELECT $1,c.video_id,c.id,TRUE,CURRENT_TIMESTAMP
		FROM video_captions c WHERE c.video_id=$2 AND c.id=$3
		ON CONFLICT(user_id,caption_id) DO UPDATE
		SET video_id=EXCLUDED.video_id,is_completed=TRUE,completed_at=CURRENT_TIMESTAMP
		RETURNING video_id,caption_id,completed_at
	)
	SELECT saved.video_id,saved.caption_id,c.sentence_order,c.start_time_ms,c.end_time_ms,c.content,saved.completed_at
	FROM saved JOIN video_captions c ON c.id=saved.caption_id AND c.video_id=saved.video_id`,
		userID, videoID, captionID).Scan(&item.VideoID, &item.CaptionID, &item.SentenceOrder,
		&item.StartTimeMS, &item.EndTimeMS, &item.Content, &item.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.DictationProgress{}, entity.ErrCaptionNotFound
	}
	if err != nil {
		return entity.DictationProgress{}, fmt.Errorf("UserRepo - CompleteDictation: %w", err)
	}
	return item, nil
}

func (r *Repo) ListCompletedDictations(ctx context.Context, userID string, videoID int64) (entity.DictationProgressList, error) {
	rows, err := r.Pool.Query(ctx, `SELECT progress.video_id,progress.caption_id,c.sentence_order,c.start_time_ms,
		c.end_time_ms,c.content,progress.completed_at,COUNT(*) OVER()
		FROM dictation_progress progress
		JOIN video_captions c ON c.id=progress.caption_id AND c.video_id=progress.video_id
		WHERE progress.user_id=$1 AND progress.video_id=$2 AND progress.is_completed=TRUE
		ORDER BY c.sentence_order ASC,c.id ASC`, userID, videoID)
	if err != nil {
		return entity.DictationProgressList{}, fmt.Errorf("UserRepo - ListCompletedDictations: %w", err)
	}
	defer rows.Close()
	result := entity.DictationProgressList{Items: make([]entity.DictationProgress, 0)}
	for rows.Next() {
		var item entity.DictationProgress
		if err = rows.Scan(&item.VideoID, &item.CaptionID, &item.SentenceOrder, &item.StartTimeMS,
			&item.EndTimeMS, &item.Content, &item.CompletedAt, &result.Total); err != nil {
			return entity.DictationProgressList{}, fmt.Errorf("UserRepo - ListCompletedDictations - scan: %w", err)
		}
		result.Items = append(result.Items, item)
	}
	if err = rows.Err(); err != nil {
		return entity.DictationProgressList{}, fmt.Errorf("UserRepo - ListCompletedDictations - rows: %w", err)
	}
	return result, nil
}

// GetByID -.
func (r *Repo) GetByID(ctx context.Context, id string) (entity.User, error) {
	return r.getUser(ctx, "id", id)
}

// GetByEmail -.
func (r *Repo) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	return r.getUser(ctx, "email", email)
}

// GetByFirebaseUID -.
func (r *Repo) GetByFirebaseUID(ctx context.Context, firebaseUID string) (entity.User, error) {
	return r.getUser(ctx, "firebase_uid", firebaseUID)
}

func (r *Repo) getUser(ctx context.Context, column, value string) (entity.User, error) {
	sql, args, err := r.Builder.
		Select("id, COALESCE(firebase_uid, ''), username, email, role, COALESCE(avatar_url, ''), native_language_id, target_language_id, is_active, created_at, updated_at").
		From("users").
		Where(sq.Eq{column: value}).
		ToSql()
	if err != nil {
		return entity.User{}, fmt.Errorf("UserRepo - getUser - r.Builder: %w", err)
	}

	var user entity.User

	err = r.Pool.QueryRow(ctx, sql, args...).
		Scan(&user.ID, &user.FirebaseUID, &user.Username, &user.Email, &user.Role, &user.AvatarURL,
			&user.NativeLanguageID, &user.TargetLanguageID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, entity.ErrUserNotFound
		}

		return entity.User{}, fmt.Errorf("UserRepo - getUser - r.Pool.QueryRow: %w", err)
	}

	return user, nil
}
