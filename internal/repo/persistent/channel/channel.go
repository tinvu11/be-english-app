// Package channel implements PostgreSQL storage for YouTube channels.
package channel

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

func New(pg *postgres.Postgres) repo.ChannelRepo { return newTraced(&Repo{Postgres: pg}) }

func (r *Repo) ListChannels(ctx context.Context) ([]entity.Channel, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id,channel_youtube_id,channel_name,COALESCE(avatar_url,''),is_active,created_at
		FROM channels ORDER BY channel_name,id`)
	if err != nil {
		return nil, fmt.Errorf("ChannelRepo - ListChannels: %w", err)
	}
	defer rows.Close()

	items := make([]entity.Channel, 0)
	for rows.Next() {
		var item entity.Channel
		if err = rows.Scan(&item.ID, &item.ChannelYouTubeID, &item.ChannelName, &item.AvatarURL, &item.IsActive, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("ChannelRepo - ListChannels - scan: %w", err)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ChannelRepo - ListChannels - rows: %w", err)
	}

	return items, nil
}

func (r *Repo) CreateChannel(ctx context.Context, channel *entity.Channel) error {
	err := r.Pool.QueryRow(ctx, `INSERT INTO channels(channel_youtube_id,channel_name,avatar_url,is_active)
		VALUES($1,$2,$3,$4) RETURNING id,created_at`, channel.ChannelYouTubeID, channel.ChannelName,
		nullableText(channel.AvatarURL), channel.IsActive).Scan(&channel.ID, &channel.CreatedAt)
	if err != nil {
		return mapWriteError("CreateChannel", err)
	}
	return nil
}

func (r *Repo) UpdateChannel(ctx context.Context, channel *entity.Channel) error {
	err := r.Pool.QueryRow(ctx, `UPDATE channels SET channel_youtube_id=$2,channel_name=$3,avatar_url=$4,is_active=$5
		WHERE id=$1 RETURNING created_at`, channel.ID, channel.ChannelYouTubeID, channel.ChannelName,
		nullableText(channel.AvatarURL), channel.IsActive).Scan(&channel.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.ErrChannelNotFound
	}
	if err != nil {
		return mapWriteError("UpdateChannel", err)
	}
	return nil
}

func (r *Repo) DeleteChannel(ctx context.Context, id int) error {
	result, err := r.Pool.Exec(ctx, `DELETE FROM channels WHERE id=$1`, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return entity.ErrChannelReferenced
		}
		return fmt.Errorf("ChannelRepo - DeleteChannel: %w", err)
	}
	if result.RowsAffected() == 0 {
		return entity.ErrChannelNotFound
	}
	return nil
}

func mapWriteError(operation string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return entity.ErrChannelExists
	}
	return fmt.Errorf("ChannelRepo - %s: %w", operation, err)
}

func nullableText(value string) any {
	if value == "" {
		return nil
	}
	return value
}
