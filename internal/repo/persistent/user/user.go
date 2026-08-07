// Package user implements the Postgres-backed User repository.
package user

import (
	"context"
	"errors"
	"fmt"

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
