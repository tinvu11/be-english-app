// Package content implements PostgreSQL persistence for the content catalog.
package content

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

const idColumn = "id"

// Repo -.
type Repo struct{ *postgres.Postgres }

// New returns a content repository instrumented with OpenTelemetry tracing spans.
func New(pg *postgres.Postgres) repo.ContentRepo { return newTraced(&Repo{pg}) }

func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return entity.ErrContentConflict
		case "23503":
			return entity.ErrInvalidReference
		}
	}

	return err
}

// CreateTopic -.
func (r *Repo) CreateTopic(ctx context.Context, item *entity.Topic) error {
	q, a, err := r.Builder.Insert("topics").Columns("name", "slug", "icon", "description", "sort_order", "is_active").
		Values(item.Name, item.Slug, item.Icon, item.Description, item.SortOrder, item.IsActive).
		Suffix("RETURNING id, created_at").ToSql()
	if err == nil {
		err = r.Pool.QueryRow(ctx, q, a...).Scan(&item.ID, &item.CreatedAt)
	}

	return wrap("CreateTopic", mapWriteError(err))
}

// GetTopic -.
func (r *Repo) GetTopic(ctx context.Context, id int64) (entity.Topic, error) {
	var item entity.Topic
	q, a, err := r.Builder.Select("id", "name", "slug", "COALESCE(icon,'')", "COALESCE(description,'')", "sort_order", "is_active", "created_at").
		From("topics").Where(sq.Eq{idColumn: id}).ToSql()
	if err == nil {
		err = r.Pool.QueryRow(ctx, q, a...).Scan(&item.ID, &item.Name, &item.Slug, &item.Icon, &item.Description, &item.SortOrder, &item.IsActive, &item.CreatedAt)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return item, entity.ErrContentNotFound
	}

	return item, wrap("GetTopic", err)
}

// ListTopics -.
func (r *Repo) ListTopics(ctx context.Context, f repo.ContentFilter) ([]entity.Topic, int, error) {
	var total int
	if err := r.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM topics").Scan(&total); err != nil {
		return nil, 0, wrap("ListTopics count", err)
	}
	q, a, err := r.Builder.Select("id", "name", "slug", "COALESCE(icon,'')", "COALESCE(description,'')", "sort_order", "is_active", "created_at").
		From("topics").OrderBy("sort_order ASC", "id ASC").Limit(f.Limit).Offset(f.Offset).ToSql()
	if err != nil {
		return nil, 0, wrap("ListTopics build", err)
	}
	rows, err := r.Pool.Query(ctx, q, a...)
	if err != nil {
		return nil, 0, wrap("ListTopics query", err)
	}
	defer rows.Close()
	items := make([]entity.Topic, 0)
	for rows.Next() {
		var item entity.Topic
		if err = rows.Scan(&item.ID, &item.Name, &item.Slug, &item.Icon, &item.Description, &item.SortOrder, &item.IsActive, &item.CreatedAt); err != nil {
			return nil, 0, wrap("ListTopics scan", err)
		}
		items = append(items, item)
	}

	return items, total, wrap("ListTopics rows", rows.Err())
}

// UpdateTopic -.
func (r *Repo) UpdateTopic(ctx context.Context, item *entity.Topic) error {
	q, a, err := r.Builder.Update("topics").Set("name", item.Name).Set("slug", item.Slug).Set("icon", item.Icon).
		Set("description", item.Description).Set("sort_order", item.SortOrder).Set("is_active", item.IsActive).
		Where(sq.Eq{idColumn: item.ID}).ToSql()

	return r.execUpdate(ctx, "UpdateTopic", q, a, err)
}

// DeleteTopic -.
func (r *Repo) DeleteTopic(ctx context.Context, id int64) error {
	return r.delete(ctx, "topics", id)
}

// CreateLevel -.
func (r *Repo) CreateLevel(ctx context.Context, item *entity.Level) error {
	q, a, err := r.Builder.Insert("levels").Columns("code", "name", "sort_order").Values(item.Code, item.Name, item.SortOrder).
		Suffix("RETURNING id, created_at").ToSql()
	if err == nil {
		err = r.Pool.QueryRow(ctx, q, a...).Scan(&item.ID, &item.CreatedAt)
	}

	return wrap("CreateLevel", mapWriteError(err))
}

// GetLevel -.
func (r *Repo) GetLevel(ctx context.Context, id int64) (entity.Level, error) {
	var item entity.Level
	q, a, err := r.Builder.Select("id", "code", "name", "sort_order", "created_at").From("levels").Where(sq.Eq{idColumn: id}).ToSql()
	if err == nil {
		err = r.Pool.QueryRow(ctx, q, a...).Scan(&item.ID, &item.Code, &item.Name, &item.SortOrder, &item.CreatedAt)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return item, entity.ErrContentNotFound
	}

	return item, wrap("GetLevel", err)
}

// ListLevels -.
func (r *Repo) ListLevels(ctx context.Context, f repo.ContentFilter) ([]entity.Level, int, error) {
	var total int
	if err := r.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM levels").Scan(&total); err != nil {
		return nil, 0, wrap("ListLevels count", err)
	}
	q, a, err := r.Builder.Select("id", "code", "name", "sort_order", "created_at").From("levels").
		OrderBy("sort_order ASC", "id ASC").Limit(f.Limit).Offset(f.Offset).ToSql()
	if err != nil {
		return nil, 0, wrap("ListLevels build", err)
	}
	rows, err := r.Pool.Query(ctx, q, a...)
	if err != nil {
		return nil, 0, wrap("ListLevels query", err)
	}
	defer rows.Close()
	items := make([]entity.Level, 0)
	for rows.Next() {
		var item entity.Level
		if err = rows.Scan(&item.ID, &item.Code, &item.Name, &item.SortOrder, &item.CreatedAt); err != nil {
			return nil, 0, wrap("ListLevels scan", err)
		}
		items = append(items, item)
	}

	return items, total, wrap("ListLevels rows", rows.Err())
}

// UpdateLevel -.
func (r *Repo) UpdateLevel(ctx context.Context, item *entity.Level) error {
	q, a, err := r.Builder.Update("levels").Set("code", item.Code).Set("name", item.Name).
		Set("sort_order", item.SortOrder).Where(sq.Eq{idColumn: item.ID}).ToSql()

	return r.execUpdate(ctx, "UpdateLevel", q, a, err)
}

// DeleteLevel -.
func (r *Repo) DeleteLevel(ctx context.Context, id int64) error {
	return r.delete(ctx, "levels", id)
}

// CreateChannel -.
func (r *Repo) CreateChannel(ctx context.Context, item *entity.Channel) error {
	q, a, err := r.Builder.Insert("channels").Columns("channel_id", "channel_name", "thumbnail_url").
		Values(item.ChannelID, item.ChannelName, item.ThumbnailURL).
		Suffix("RETURNING id, created_at, updated_at").ToSql()
	if err == nil {
		err = r.Pool.QueryRow(ctx, q, a...).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	}

	return wrap("CreateChannel", mapWriteError(err))
}

// GetChannel -.
func (r *Repo) GetChannel(ctx context.Context, id int64) (entity.Channel, error) {
	var item entity.Channel
	q, a, err := r.Builder.Select("id", "channel_id", "channel_name", "COALESCE(thumbnail_url,'')", "created_at", "updated_at").
		From("channels").Where(sq.Eq{idColumn: id}).ToSql()
	if err == nil {
		err = r.Pool.QueryRow(ctx, q, a...).Scan(&item.ID, &item.ChannelID, &item.ChannelName, &item.ThumbnailURL, &item.CreatedAt, &item.UpdatedAt)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return item, entity.ErrContentNotFound
	}

	return item, wrap("GetChannel", err)
}

// ListChannels -.
func (r *Repo) ListChannels(ctx context.Context, f repo.ContentFilter) ([]entity.Channel, int, error) {
	var total int
	if err := r.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM channels").Scan(&total); err != nil {
		return nil, 0, wrap("ListChannels count", err)
	}
	q, a, err := r.Builder.Select("id", "channel_id", "channel_name", "COALESCE(thumbnail_url,'')", "created_at", "updated_at").
		From("channels").OrderBy("id ASC").Limit(f.Limit).Offset(f.Offset).ToSql()
	if err != nil {
		return nil, 0, wrap("ListChannels build", err)
	}
	rows, err := r.Pool.Query(ctx, q, a...)
	if err != nil {
		return nil, 0, wrap("ListChannels query", err)
	}
	defer rows.Close()
	items := make([]entity.Channel, 0)
	for rows.Next() {
		var item entity.Channel
		if err = rows.Scan(&item.ID, &item.ChannelID, &item.ChannelName, &item.ThumbnailURL, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, wrap("ListChannels scan", err)
		}
		items = append(items, item)
	}

	return items, total, wrap("ListChannels rows", rows.Err())
}

// UpdateChannel -.
func (r *Repo) UpdateChannel(ctx context.Context, item *entity.Channel) error {
	q, a, err := r.Builder.Update("channels").Set("channel_id", item.ChannelID).Set("channel_name", item.ChannelName).
		Set("thumbnail_url", item.ThumbnailURL).Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{idColumn: item.ID}).Suffix("RETURNING updated_at").ToSql()
	if err == nil {
		err = r.Pool.QueryRow(ctx, q, a...).Scan(&item.UpdatedAt)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.ErrContentNotFound
	}

	return wrap("UpdateChannel", mapWriteError(err))
}

// DeleteChannel -.
func (r *Repo) DeleteChannel(ctx context.Context, id int64) error {
	return r.delete(ctx, "channels", id)
}

// CreateVideo -.
func (r *Repo) CreateVideo(ctx context.Context, item *entity.Video) error {
	return r.writeVideo(ctx, item, true)
}

// GetVideo -.
func (r *Repo) GetVideo(ctx context.Context, id int64) (entity.Video, error) {
	var item entity.Video
	q, a, err := r.Builder.Select(
		"v.id", "v.video_id", "v.channel_id", "v.title", "COALESCE(v.description,'')",
		"COALESCE(v.thumbnail_url,'')", "v.view_count", "v.duration", "v.sort_order", "v.is_active",
		"v.published_at", "v.created_at", "v.updated_at",
		"COALESCE(array_agg(DISTINCT vt.topic_id) FILTER (WHERE vt.topic_id IS NOT NULL), '{}')",
		"COALESCE(array_agg(DISTINCT vl.level_id::bigint) FILTER (WHERE vl.level_id IS NOT NULL), '{}')",
	).From("videos v").LeftJoin("video_topics vt ON vt.video_id = v.id").
		LeftJoin("video_levels vl ON vl.video_id = v.id").Where(sq.Eq{"v.id": id}).GroupBy("v.id").ToSql()
	if err == nil {
		err = scanVideo(r.Pool.QueryRow(ctx, q, a...), &item)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return item, entity.ErrContentNotFound
	}

	return item, wrap("GetVideo", err)
}

// ListVideos -.
func (r *Repo) ListVideos(ctx context.Context, f repo.ContentFilter) ([]entity.Video, int, error) {
	var total int
	if err := r.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM videos").Scan(&total); err != nil {
		return nil, 0, wrap("ListVideos count", err)
	}
	q, a, err := r.Builder.Select(
		"v.id", "v.video_id", "v.channel_id", "v.title", "COALESCE(v.description,'')",
		"COALESCE(v.thumbnail_url,'')", "v.view_count", "v.duration", "v.sort_order", "v.is_active",
		"v.published_at", "v.created_at", "v.updated_at",
		"COALESCE(array_agg(DISTINCT vt.topic_id) FILTER (WHERE vt.topic_id IS NOT NULL), '{}')",
		"COALESCE(array_agg(DISTINCT vl.level_id::bigint) FILTER (WHERE vl.level_id IS NOT NULL), '{}')",
	).From("videos v").LeftJoin("video_topics vt ON vt.video_id = v.id").
		LeftJoin("video_levels vl ON vl.video_id = v.id").GroupBy("v.id").
		OrderBy("v.sort_order ASC", "v.id ASC").Limit(f.Limit).Offset(f.Offset).ToSql()
	if err != nil {
		return nil, 0, wrap("ListVideos build", err)
	}
	rows, err := r.Pool.Query(ctx, q, a...)
	if err != nil {
		return nil, 0, wrap("ListVideos query", err)
	}
	defer rows.Close()
	items := make([]entity.Video, 0)
	for rows.Next() {
		var item entity.Video
		if err = scanVideo(rows, &item); err != nil {
			return nil, 0, wrap("ListVideos scan", err)
		}
		items = append(items, item)
	}

	return items, total, wrap("ListVideos rows", rows.Err())
}

// UpdateVideo -.
func (r *Repo) UpdateVideo(ctx context.Context, item *entity.Video) error {
	return r.writeVideo(ctx, item, false)
}

// DeleteVideo -.
func (r *Repo) DeleteVideo(ctx context.Context, id int64) error {
	return r.delete(ctx, "videos", id)
}

func (r *Repo) writeVideo(ctx context.Context, item *entity.Video, create bool) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return wrap("writeVideo begin", err)
	}
	defer rollback(ctx, tx)

	if create {
		q, a, buildErr := r.Builder.Insert("videos").
			Columns("video_id", "channel_id", "title", "description", "thumbnail_url", "view_count", "duration", "sort_order", "is_active", "published_at").
			Values(item.VideoID, item.ChannelID, item.Title, item.Description, item.ThumbnailURL, item.ViewCount, item.Duration, item.SortOrder, item.IsActive, item.PublishedAt).
			Suffix("RETURNING id, created_at, updated_at").ToSql()
		if buildErr != nil {
			return wrap("writeVideo build insert", buildErr)
		}
		err = tx.QueryRow(ctx, q, a...).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	} else {
		q, a, buildErr := r.Builder.Update("videos").Set("video_id", item.VideoID).Set("channel_id", item.ChannelID).
			Set("title", item.Title).Set("description", item.Description).Set("thumbnail_url", item.ThumbnailURL).
			Set("view_count", item.ViewCount).Set("duration", item.Duration).Set("sort_order", item.SortOrder).
			Set("is_active", item.IsActive).Set("published_at", item.PublishedAt).Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
			Where(sq.Eq{idColumn: item.ID}).Suffix("RETURNING created_at, updated_at").ToSql()
		if buildErr != nil {
			return wrap("writeVideo build update", buildErr)
		}
		err = tx.QueryRow(ctx, q, a...).Scan(&item.CreatedAt, &item.UpdatedAt)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.ErrContentNotFound
	}
	if err != nil {
		return wrap("writeVideo write", mapWriteError(err))
	}
	if !create {
		if _, err = tx.Exec(ctx, "DELETE FROM video_topics WHERE video_id = $1", item.ID); err != nil {
			return wrap("writeVideo clear topics", err)
		}
		if _, err = tx.Exec(ctx, "DELETE FROM video_levels WHERE video_id = $1", item.ID); err != nil {
			return wrap("writeVideo clear levels", err)
		}
	}
	for _, topicID := range item.TopicIDs {
		if _, err = tx.Exec(ctx, "INSERT INTO video_topics(video_id, topic_id) VALUES ($1, $2)", item.ID, topicID); err != nil {
			return wrap("writeVideo topic", mapWriteError(err))
		}
	}
	for _, levelID := range item.LevelIDs {
		if _, err = tx.Exec(ctx, "INSERT INTO video_levels(video_id, level_id) VALUES ($1, $2)", item.ID, levelID); err != nil {
			return wrap("writeVideo level", mapWriteError(err))
		}
	}

	return wrap("writeVideo commit", tx.Commit(ctx))
}

func scanVideo(row pgx.Row, item *entity.Video) error {
	return row.Scan(
		&item.ID, &item.VideoID, &item.ChannelID, &item.Title, &item.Description, &item.ThumbnailURL,
		&item.ViewCount, &item.Duration, &item.SortOrder, &item.IsActive, &item.PublishedAt,
		&item.CreatedAt, &item.UpdatedAt, &item.TopicIDs, &item.LevelIDs,
	)
}

func rollback(ctx context.Context, tx pgx.Tx) {
	err := tx.Rollback(ctx)
	if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		return
	}
}

func (r *Repo) execUpdate(ctx context.Context, op, query string, args []any, buildErr error) error {
	if buildErr != nil {
		return wrap(op, buildErr)
	}
	result, err := r.Pool.Exec(ctx, query, args...)
	if err != nil {
		return wrap(op, mapWriteError(err))
	}
	if result.RowsAffected() == 0 {
		return entity.ErrContentNotFound
	}

	return nil
}

func (r *Repo) delete(ctx context.Context, table string, id int64) error {
	q, a, err := r.Builder.Delete(table).Where(sq.Eq{idColumn: id}).ToSql()
	if err != nil {
		return wrap("delete", err)
	}
	result, err := r.Pool.Exec(ctx, q, a...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return entity.ErrContentReferenced
		}

		return wrap("delete", err)
	}
	if result.RowsAffected() == 0 {
		return entity.ErrContentNotFound
	}

	return nil
}

func wrap(op string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("ContentRepo - %s: %w", op, err)
}
