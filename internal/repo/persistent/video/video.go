// Package video implements PostgreSQL storage for videos.
package video

import (
	"context"
	"errors"
	"fmt"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/goccy/go-json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const videoColumns = `v.id,v.title,v.youtube_id,COALESCE(v.thumbnail_url,''),v.duration_seconds,v.status,
	v.created_at,v.updated_at,
	lang.id,lang.code,lang.name,
	COALESCE(l.id,0),COALESCE(l.code,''),COALESCE((SELECT lt.name FROM level_translations lt WHERE lt.level_id=l.id
		ORDER BY (lt.language_id=v.language_id) DESC,lt.language_id LIMIT 1),l.code,''),
	COALESCE(c.id,0),COALESCE(c.channel_youtube_id,''),COALESCE(c.channel_name,''),COALESCE(c.avatar_url,''),
	COALESCE((SELECT jsonb_agg(jsonb_build_object(
		'id',t.id,'slug',t.slug,'name',COALESCE((SELECT tt.name FROM topic_translations tt
			WHERE tt.topic_id=t.id ORDER BY (tt.language_id=v.language_id) DESC,tt.language_id LIMIT 1),t.slug),
		'iconUrl',COALESCE(t.icon_url,'')) ORDER BY t.slug)
		FROM video_topics vt JOIN topics t ON t.id=vt.topic_id WHERE vt.video_id=v.id),'[]'::jsonb),
	(SELECT COUNT(*) FROM video_captions vc WHERE vc.video_id=v.id),
	COALESCE((SELECT jsonb_agg(jsonb_build_object(
		'languageId',translated.language_id,'languageCode',translated.language_code,
		'languageName',translated.language_name,'translatedCaptionCount',translated.caption_count,
		'isComplete',translated.caption_count=(SELECT COUNT(*) FROM video_captions all_captions WHERE all_captions.video_id=v.id))
		ORDER BY translated.language_code)
		FROM (SELECT lang_translation.id AS language_id,lang_translation.code AS language_code,
			lang_translation.name AS language_name,COUNT(*) AS caption_count
			FROM video_captions translated_caption
			JOIN caption_translations ct ON ct.caption_id=translated_caption.id
			JOIN languages lang_translation ON lang_translation.id=ct.language_id
			WHERE translated_caption.video_id=v.id
			GROUP BY lang_translation.id,lang_translation.code,lang_translation.name) translated),'[]'::jsonb)`

const videoJoins = ` FROM videos v
	JOIN languages lang ON lang.id=v.language_id
	LEFT JOIN levels l ON l.id=v.level_id
	LEFT JOIN channels c ON c.id=v.channel_id `

type Repo struct{ *postgres.Postgres }

func New(pg *postgres.Postgres) repo.VideoRepo { return newTraced(&Repo{Postgres: pg}) }

func (r *Repo) ListVideos(ctx context.Context, filter entity.VideoFilter) (entity.VideoList, error) {
	query := `SELECT ` + videoColumns + `,COUNT(*) OVER()` + videoJoins + `
		WHERE v.is_system=TRUE
		AND ($1::int IS NULL OR v.language_id=$1)
		AND ($2::int IS NULL OR v.level_id=$2)
		AND ($3::int IS NULL OR v.channel_id=$3)
		AND ($4='' OR v.status=$4)
		AND ($5='' OR v.title ILIKE '%' || $5 || '%')
		ORDER BY v.created_at DESC,v.id DESC LIMIT $6 OFFSET $7`
	rows, err := r.Pool.Query(ctx, query, filter.LanguageID, filter.LevelID, filter.ChannelID,
		filter.Status, filter.Search, filter.Limit, filter.Offset)
	if err != nil {
		return entity.VideoList{}, fmt.Errorf("VideoRepo - ListVideos: %w", err)
	}
	defer rows.Close()

	result := entity.VideoList{Items: make([]entity.Video, 0)}
	for rows.Next() {
		video, scanErr := scanVideo(rows, &result.Total)
		if scanErr != nil {
			return entity.VideoList{}, fmt.Errorf("VideoRepo - ListVideos - scan: %w", scanErr)
		}
		result.Items = append(result.Items, video)
	}
	if err = rows.Err(); err != nil {
		return entity.VideoList{}, fmt.Errorf("VideoRepo - ListVideos - rows: %w", err)
	}
	return result, nil
}

func (r *Repo) GetVideo(ctx context.Context, id int64) (entity.Video, error) {
	video, err := scanVideo(r.Pool.QueryRow(ctx, `SELECT `+videoColumns+videoJoins+` WHERE v.id=$1`, id), nil)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.Video{}, entity.ErrVideoNotFound
	}
	if err != nil {
		return entity.Video{}, fmt.Errorf("VideoRepo - GetVideo: %w", err)
	}
	return video, nil
}

func (r *Repo) CreateVideo(ctx context.Context, input entity.VideoInput) (entity.Video, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return entity.Video{}, fmt.Errorf("VideoRepo - CreateVideo - begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var id int64
	err = tx.QueryRow(ctx, `INSERT INTO videos(title,youtube_id,thumbnail_url,duration_seconds,status,language_id,level_id,channel_id,is_system)
		VALUES($1,$2,NULLIF($3,''),$4,$5,$6,$7,$8,TRUE) RETURNING id`, input.Title, input.YouTubeID,
		input.ThumbnailURL, input.DurationSeconds, input.Status, input.LanguageID, input.LevelID, input.ChannelID).Scan(&id)
	if err != nil {
		return entity.Video{}, mapWriteError("CreateVideo", err)
	}
	if err = replaceTopics(ctx, tx, id, input.TopicIDs); err != nil {
		return entity.Video{}, mapWriteError("CreateVideo topics", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return entity.Video{}, fmt.Errorf("VideoRepo - CreateVideo - commit: %w", err)
	}
	return r.GetVideo(ctx, id)
}

func (r *Repo) UpdateVideo(ctx context.Context, id int64, input entity.VideoInput) (entity.Video, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return entity.Video{}, fmt.Errorf("VideoRepo - UpdateVideo - begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	result, err := tx.Exec(ctx, `UPDATE videos SET title=$2,youtube_id=$3,thumbnail_url=NULLIF($4,''),duration_seconds=$5,
		status=$6,language_id=$7,level_id=$8,channel_id=$9,updated_at=CURRENT_TIMESTAMP WHERE id=$1`,
		id, input.Title, input.YouTubeID, input.ThumbnailURL, input.DurationSeconds, input.Status,
		input.LanguageID, input.LevelID, input.ChannelID)
	if err != nil {
		return entity.Video{}, mapWriteError("UpdateVideo", err)
	}
	if result.RowsAffected() == 0 {
		return entity.Video{}, entity.ErrVideoNotFound
	}
	if _, err = tx.Exec(ctx, `DELETE FROM video_topics WHERE video_id=$1`, id); err != nil {
		return entity.Video{}, fmt.Errorf("VideoRepo - UpdateVideo - delete topics: %w", err)
	}
	if err = replaceTopics(ctx, tx, id, input.TopicIDs); err != nil {
		return entity.Video{}, mapWriteError("UpdateVideo topics", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return entity.Video{}, fmt.Errorf("VideoRepo - UpdateVideo - commit: %w", err)
	}
	return r.GetVideo(ctx, id)
}

func (r *Repo) SetVideoStatus(ctx context.Context, id int64, expectedStatus, nextStatus string) (entity.Video, error) {
	result, err := r.Pool.Exec(ctx, `UPDATE videos SET status=$3,updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND status=$2`,
		id, expectedStatus, nextStatus)
	if err != nil {
		return entity.Video{}, fmt.Errorf("VideoRepo - SetVideoStatus: %w", err)
	}
	if result.RowsAffected() == 0 {
		if _, getErr := r.GetVideo(ctx, id); errors.Is(getErr, entity.ErrVideoNotFound) {
			return entity.Video{}, entity.ErrVideoNotFound
		}
		return entity.Video{}, entity.ErrConcurrentVideoUpdate
	}
	return r.GetVideo(ctx, id)
}

func (r *Repo) DeleteVideo(ctx context.Context, id int64) error {
	result, err := r.Pool.Exec(ctx, `DELETE FROM videos WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("VideoRepo - DeleteVideo: %w", err)
	}
	if result.RowsAffected() == 0 {
		return entity.ErrVideoNotFound
	}
	return nil
}

type scanner interface{ Scan(...any) error }

func scanVideo(row scanner, total *int) (entity.Video, error) {
	var video entity.Video
	var topicsJSON []byte
	var translationsJSON []byte
	destinations := []any{&video.ID, &video.Title, &video.YouTubeID, &video.ThumbnailURL, &video.DurationSeconds,
		&video.Status, &video.CreatedAt, &video.UpdatedAt, &video.Language.ID, &video.Language.Code,
		&video.Language.Name, &video.Level.ID, &video.Level.Code, &video.Level.Name, &video.Channel.ID,
		&video.Channel.ChannelYouTubeID, &video.Channel.Name, &video.Channel.AvatarURL, &topicsJSON,
		&video.CaptionAvailability.CaptionCount, &translationsJSON}
	if total != nil {
		destinations = append(destinations, total)
	}
	if err := row.Scan(destinations...); err != nil {
		return video, err
	}
	if err := json.Unmarshal(topicsJSON, &video.Topics); err != nil {
		return video, fmt.Errorf("decode topics: %w", err)
	}
	video.CaptionAvailability.HasOriginal = video.CaptionAvailability.CaptionCount > 0
	if err := json.Unmarshal(translationsJSON, &video.CaptionAvailability.Translations); err != nil {
		return video, fmt.Errorf("decode caption availability: %w", err)
	}
	video.VideoURL = "https://www.youtube.com/watch?v=" + video.YouTubeID
	return video, nil
}

func replaceTopics(ctx context.Context, tx pgx.Tx, videoID int64, topicIDs []int) error {
	for _, topicID := range topicIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO video_topics(video_id,topic_id) VALUES($1,$2)`, videoID, topicID); err != nil {
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
			return entity.ErrVideoExists
		case "23503":
			return entity.ErrInvalidReference
		}
	}
	return fmt.Errorf("VideoRepo - %s: %w", operation, err)
}
