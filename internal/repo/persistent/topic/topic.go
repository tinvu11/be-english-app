// Package topic implements PostgreSQL storage for topics and translations.
package topic

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

type Repo struct{ *postgres.Postgres }

func New(pg *postgres.Postgres) repo.TopicRepo { return newTraced(&Repo{Postgres: pg}) }

func (r *Repo) ListTopics(ctx context.Context) ([]entity.Topic, error) {
	rows, err := r.Pool.Query(ctx, `SELECT t.id,t.slug,COALESCE(t.icon_url,''),t.is_active,t.created_at,
		tt.language_id,tt.name FROM topics t LEFT JOIN topic_translations tt ON tt.topic_id=t.id
		ORDER BY t.slug,t.id,tt.language_id`)
	if err != nil {
		return nil, fmt.Errorf("TopicRepo - ListTopics: %w", err)
	}
	defer rows.Close()

	items := make([]entity.Topic, 0)
	positions := make(map[int]int)
	for rows.Next() {
		var item entity.Topic
		var languageID *int
		var name *string
		if err = rows.Scan(&item.ID, &item.Slug, &item.IconURL, &item.IsActive, &item.CreatedAt, &languageID, &name); err != nil {
			return nil, fmt.Errorf("TopicRepo - ListTopics - scan: %w", err)
		}
		position, exists := positions[item.ID]
		if !exists {
			position = len(items)
			positions[item.ID] = position
			item.Translations = make([]entity.TopicTranslation, 0)
			items = append(items, item)
		}
		if languageID != nil && name != nil {
			items[position].Translations = append(items[position].Translations, entity.TopicTranslation{LanguageID: *languageID, Name: *name})
		}
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("TopicRepo - ListTopics - rows: %w", err)
	}

	return items, nil
}

func (r *Repo) CreateTopic(ctx context.Context, topic *entity.Topic) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("TopicRepo - CreateTopic - begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	err = tx.QueryRow(ctx, `INSERT INTO topics(slug,icon_url,is_active) VALUES($1,$2,$3) RETURNING id,created_at`,
		topic.Slug, nullableText(topic.IconURL), topic.IsActive).Scan(&topic.ID, &topic.CreatedAt)
	if err != nil {
		return mapWriteError("CreateTopic", err)
	}
	if err = insertTranslations(ctx, tx, topic.ID, topic.Translations); err != nil {
		return mapWriteError("CreateTopic translations", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("TopicRepo - CreateTopic - commit: %w", err)
	}

	return nil
}

func (r *Repo) UpdateTopic(ctx context.Context, topic *entity.Topic) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("TopicRepo - UpdateTopic - begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	err = tx.QueryRow(ctx, `UPDATE topics SET slug=$2,icon_url=$3,is_active=$4 WHERE id=$1 RETURNING created_at`,
		topic.ID, topic.Slug, nullableText(topic.IconURL), topic.IsActive).Scan(&topic.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.ErrTopicNotFound
	}
	if err != nil {
		return mapWriteError("UpdateTopic", err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM topic_translations WHERE topic_id=$1`, topic.ID); err != nil {
		return fmt.Errorf("TopicRepo - UpdateTopic - delete translations: %w", err)
	}
	if err = insertTranslations(ctx, tx, topic.ID, topic.Translations); err != nil {
		return mapWriteError("UpdateTopic translations", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("TopicRepo - UpdateTopic - commit: %w", err)
	}

	return nil
}

func (r *Repo) SetTopicActive(ctx context.Context, id int, isActive bool) (entity.Topic, error) {
	var topic entity.Topic
	err := r.Pool.QueryRow(ctx, `UPDATE topics SET is_active=$2 WHERE id=$1
		RETURNING id,slug,COALESCE(icon_url,''),is_active,created_at`, id, isActive).
		Scan(&topic.ID, &topic.Slug, &topic.IconURL, &topic.IsActive, &topic.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return topic, entity.ErrTopicNotFound
	}
	if err != nil {
		return topic, fmt.Errorf("TopicRepo - SetTopicActive: %w", err)
	}
	topic.Translations = make([]entity.TopicTranslation, 0)

	return topic, nil
}

func (r *Repo) DeleteTopic(ctx context.Context, id int) error {
	result, err := r.Pool.Exec(ctx, `DELETE FROM topics WHERE id=$1`, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return entity.ErrTopicReferenced
		}
		return fmt.Errorf("TopicRepo - DeleteTopic: %w", err)
	}
	if result.RowsAffected() == 0 {
		return entity.ErrTopicNotFound
	}

	return nil
}

func insertTranslations(ctx context.Context, tx pgx.Tx, topicID int, translations []entity.TopicTranslation) error {
	for _, translation := range translations {
		if _, err := tx.Exec(ctx, `INSERT INTO topic_translations(topic_id,language_id,name) VALUES($1,$2,$3)`,
			topicID, translation.LanguageID, translation.Name); err != nil {
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
			return entity.ErrTopicExists
		case "23503":
			return entity.ErrInvalidReference
		}
	}
	return fmt.Errorf("TopicRepo - %s: %w", operation, err)
}

func nullableText(value string) any {
	if value == "" {
		return nil
	}
	return value
}
