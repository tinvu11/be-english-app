// Package language implements PostgreSQL storage for languages.
package language

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

func New(pg *postgres.Postgres) repo.LanguageRepo {
	return newTraced(&Repo{Postgres: pg})
}

func (r *Repo) ListLanguages(ctx context.Context) ([]entity.Language, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id, code, name, is_active, created_at FROM languages ORDER BY code`)
	if err != nil {
		return nil, fmt.Errorf("LanguageRepo - ListLanguages: %w", err)
	}
	defer rows.Close()

	languages := make([]entity.Language, 0)
	for rows.Next() {
		var language entity.Language
		if err = rows.Scan(&language.ID, &language.Code, &language.Name, &language.IsActive, &language.CreatedAt); err != nil {
			return nil, fmt.Errorf("LanguageRepo - ListLanguages - scan: %w", err)
		}
		languages = append(languages, language)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("LanguageRepo - ListLanguages - rows: %w", err)
	}

	return languages, nil
}

func (r *Repo) GetLanguage(ctx context.Context, id int) (entity.Language, error) {
	var language entity.Language
	err := r.Pool.QueryRow(ctx, `SELECT id, code, name, is_active, created_at FROM languages WHERE id=$1`, id).
		Scan(&language.ID, &language.Code, &language.Name, &language.IsActive, &language.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.Language{}, entity.ErrLanguageNotFound
	}
	if err != nil {
		return entity.Language{}, fmt.Errorf("LanguageRepo - GetLanguage: %w", err)
	}
	return language, nil
}

func (r *Repo) CreateLanguage(ctx context.Context, language *entity.Language) error {
	err := r.Pool.QueryRow(ctx,
		`INSERT INTO languages(code, name, is_active) VALUES($1, $2, $3)
		 RETURNING id, created_at`,
		language.Code, language.Name, language.IsActive,
	).Scan(&language.ID, &language.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return entity.ErrLanguageExists
		}

		return fmt.Errorf("LanguageRepo - CreateLanguage: %w", err)
	}

	return nil
}

func (r *Repo) UpdateLanguage(ctx context.Context, language *entity.Language) error {
	err := r.Pool.QueryRow(ctx,
		`UPDATE languages SET name=$2, is_active=$3 WHERE id=$1
		 RETURNING code, created_at`,
		language.ID, language.Name, language.IsActive,
	).Scan(&language.Code, &language.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.ErrLanguageNotFound
		}

		return fmt.Errorf("LanguageRepo - UpdateLanguage: %w", err)
	}

	return nil
}
