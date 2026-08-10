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
