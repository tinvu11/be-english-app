// Package adminuser implements administrator-facing PostgreSQL user operations.
package adminuser

import (
	"context"
	"errors"
	"fmt"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/jackc/pgx/v5"
)

type Repo struct{ *postgres.Postgres }

func New(pg *postgres.Postgres) repo.AdminUserRepo { return newTraced(&Repo{Postgres: pg}) }

func (r *Repo) ListUsers(ctx context.Context, filter entity.UserFilter) (entity.UserList, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id,COALESCE(firebase_uid,''),username,email,role,
		COALESCE(avatar_url,''),native_language_id,target_language_id,is_active,created_at,updated_at,
		COUNT(*) OVER()
		FROM users
		WHERE ($1='' OR email ILIKE '%' || $1 || '%' OR username ILIKE '%' || $1 || '%')
		AND ($2='' OR role=$2)
		AND ($3::boolean IS NULL OR is_active=$3)
		ORDER BY created_at DESC,id
		LIMIT $4 OFFSET $5`, filter.Search, filter.Role, filter.IsActive, filter.Limit, filter.Offset)
	if err != nil {
		return entity.UserList{}, fmt.Errorf("AdminUserRepo - ListUsers: %w", err)
	}
	defer rows.Close()

	result := entity.UserList{Items: make([]entity.User, 0)}
	for rows.Next() {
		var user entity.User
		if err = rows.Scan(&user.ID, &user.FirebaseUID, &user.Username, &user.Email, &user.Role,
			&user.AvatarURL, &user.NativeLanguageID, &user.TargetLanguageID, &user.IsActive,
			&user.CreatedAt, &user.UpdatedAt, &result.Total); err != nil {
			return entity.UserList{}, fmt.Errorf("AdminUserRepo - ListUsers - scan: %w", err)
		}
		result.Items = append(result.Items, user)
	}
	if err = rows.Err(); err != nil {
		return entity.UserList{}, fmt.Errorf("AdminUserRepo - ListUsers - rows: %w", err)
	}

	return result, nil
}

func (r *Repo) SetUserActive(ctx context.Context, id string, isActive bool) (entity.User, error) {
	return r.updateUser(ctx, `UPDATE users SET is_active=$2,updated_at=CURRENT_TIMESTAMP WHERE id=$1
		RETURNING id,COALESCE(firebase_uid,''),username,email,role,COALESCE(avatar_url,''),
		native_language_id,target_language_id,is_active,created_at,updated_at`, id, isActive)
}

func (r *Repo) SetUserRole(ctx context.Context, id, role string) (entity.User, error) {
	return r.updateUser(ctx, `UPDATE users SET role=$2,updated_at=CURRENT_TIMESTAMP WHERE id=$1
		RETURNING id,COALESCE(firebase_uid,''),username,email,role,COALESCE(avatar_url,''),
		native_language_id,target_language_id,is_active,created_at,updated_at`, id, role)
}

func (r *Repo) updateUser(ctx context.Context, query string, args ...any) (entity.User, error) {
	var user entity.User
	err := r.Pool.QueryRow(ctx, query, args...).Scan(&user.ID, &user.FirebaseUID, &user.Username, &user.Email,
		&user.Role, &user.AvatarURL, &user.NativeLanguageID, &user.TargetLanguageID, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return user, entity.ErrUserNotFound
	}
	if err != nil {
		return user, fmt.Errorf("AdminUserRepo - updateUser: %w", err)
	}

	return user, nil
}
