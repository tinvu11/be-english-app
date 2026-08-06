// Package level implements PostgreSQL storage for levels and translations.
package level

import (
	"context"
	"errors"
	"fmt"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repo struct {
	*postgres.Postgres
}

func New(pg *postgres.Postgres) repo.LevelRepo {
	return newTraced(&Repo{Postgres: pg})
}

func (r *Repo) ListLevels(ctx context.Context, languageID *int) ([]entity.Level, error) {
	rows, err := r.Pool.Query(ctx, `SELECT l.id, l.code, l.language_id, l.created_at,
		lt.language_id, lt.name
		FROM levels l
		LEFT JOIN level_translations lt ON lt.level_id = l.id
		WHERE ($1::int IS NULL OR l.language_id = $1)
		ORDER BY l.code, l.id, lt.language_id`, languageID)
	if err != nil {
		return nil, fmt.Errorf("LevelRepo - ListLevels: %w", err)
	}
	defer rows.Close()

	levels := make([]entity.Level, 0)
	positions := make(map[int]int)
	for rows.Next() {
		var current entity.Level
		var translationLanguageID *int
		var translationName *string
		if err = rows.Scan(&current.ID, &current.Code, &current.LanguageID, &current.CreatedAt,
			&translationLanguageID, &translationName); err != nil {
			return nil, fmt.Errorf("LevelRepo - ListLevels - scan: %w", err)
		}

		position, exists := positions[current.ID]
		if !exists {
			position = len(levels)
			positions[current.ID] = position
			current.Translations = make([]entity.LevelTranslation, 0)
			levels = append(levels, current)
		}
		if translationLanguageID != nil && translationName != nil {
			levels[position].Translations = append(levels[position].Translations, entity.LevelTranslation{
				LanguageID: *translationLanguageID,
				Name:       *translationName,
			})
		}
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("LevelRepo - ListLevels - rows: %w", err)
	}

	return levels, nil
}

func (r *Repo) CreateLevel(ctx context.Context, level *entity.Level) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("LevelRepo - CreateLevel - begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	err = tx.QueryRow(ctx,
		`INSERT INTO levels(code, language_id) VALUES($1, $2) RETURNING id, created_at`,
		level.Code, level.LanguageID,
	).Scan(&level.ID, &level.CreatedAt)
	if err != nil {
		return mapWriteError("CreateLevel", err)
	}
	if err = replaceTranslations(ctx, tx, level.ID, level.Translations); err != nil {
		return mapWriteError("CreateLevel translations", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("LevelRepo - CreateLevel - commit: %w", err)
	}

	return nil
}

func (r *Repo) UpdateLevel(ctx context.Context, level *entity.Level) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("LevelRepo - UpdateLevel - begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	err = tx.QueryRow(ctx,
		`UPDATE levels SET code=$2, language_id=$3 WHERE id=$1 RETURNING created_at`,
		level.ID, level.Code, level.LanguageID,
	).Scan(&level.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.ErrLevelNotFound
	}
	if err != nil {
		return mapWriteError("UpdateLevel", err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM level_translations WHERE level_id=$1`, level.ID); err != nil {
		return fmt.Errorf("LevelRepo - UpdateLevel - delete translations: %w", err)
	}
	if err = replaceTranslations(ctx, tx, level.ID, level.Translations); err != nil {
		return mapWriteError("UpdateLevel translations", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("LevelRepo - UpdateLevel - commit: %w", err)
	}

	return nil
}

func (r *Repo) DeleteLevel(ctx context.Context, id int) error {
	result, err := r.Pool.Exec(ctx, `DELETE FROM levels WHERE id=$1`, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return entity.ErrLevelReferenced
		}

		return fmt.Errorf("LevelRepo - DeleteLevel: %w", err)
	}
	if result.RowsAffected() == 0 {
		return entity.ErrLevelNotFound
	}

	return nil
}

func replaceTranslations(ctx context.Context, tx pgx.Tx, levelID int, translations []entity.LevelTranslation) error {
	for _, translation := range translations {
		if _, err := tx.Exec(ctx,
			`INSERT INTO level_translations(level_id, language_id, name) VALUES($1, $2, $3)`,
			levelID, translation.LanguageID, translation.Name,
		); err != nil {
			return err
		}
	}

	return nil
}

func mapWriteError(operation string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return entity.ErrLevelExists
		case "23503":
			if pgErr.ConstraintName == "fk_videos_level_language" {
				return entity.ErrLevelReferenced
			}
			return entity.ErrInvalidReference
		}
	}

	return fmt.Errorf("LevelRepo - %s: %w", operation, err)
}
