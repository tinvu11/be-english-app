package caption

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

const captionColumns = `c.id,c.video_id,c.sentence_order,c.start_time_ms,c.end_time_ms,c.content,
	COALESCE(c.pinyin_or_furigana,''),c.created_at,
	COALESCE((SELECT jsonb_agg(jsonb_build_object('languageId',ct.language_id,'languageCode',l.code,
		'languageName',l.name,'text',ct.translated_text) ORDER BY l.code)
		FROM caption_translations ct JOIN languages l ON l.id=ct.language_id WHERE ct.caption_id=c.id),'[]'::jsonb)`

type Repo struct{ *postgres.Postgres }

func New(pg *postgres.Postgres) repo.CaptionRepo { return newTraced(&Repo{Postgres: pg}) }

func (r *Repo) ListCaptions(ctx context.Context, videoID int64, filter entity.CaptionFilter) (entity.CaptionList, error) {
	if err := r.ensureVideo(ctx, videoID); err != nil {
		return entity.CaptionList{}, err
	}
	var total int
	if err := r.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM video_captions WHERE video_id=$1`, videoID).Scan(&total); err != nil {
		return entity.CaptionList{}, fmt.Errorf("CaptionRepo - ListCaptions - count: %w", err)
	}
	rows, err := r.Pool.Query(ctx, `SELECT `+captionColumns+` FROM video_captions c WHERE c.video_id=$1
		ORDER BY c.sentence_order ASC LIMIT $2 OFFSET $3`, videoID, filter.Limit, filter.Offset)
	if err != nil {
		return entity.CaptionList{}, fmt.Errorf("CaptionRepo - ListCaptions: %w", err)
	}
	defer rows.Close()
	items := make([]entity.Caption, 0)
	for rows.Next() {
		item, scanErr := scanCaption(rows)
		if scanErr != nil {
			return entity.CaptionList{}, fmt.Errorf("CaptionRepo - ListCaptions - scan: %w", scanErr)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return entity.CaptionList{}, fmt.Errorf("CaptionRepo - ListCaptions - rows: %w", err)
	}
	return entity.CaptionList{Items: items, Total: total}, nil
}

func (r *Repo) CreateCaption(ctx context.Context, videoID int64, input entity.CaptionInput) (entity.Caption, error) {
	items, err := r.ImportCaptions(ctx, videoID, []entity.CaptionInput{input})
	if err != nil {
		return entity.Caption{}, err
	}
	return items[0], nil
}

func (r *Repo) ImportCaptions(ctx context.Context, videoID int64, inputs []entity.CaptionInput) ([]entity.Caption, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("CaptionRepo - ImportCaptions - begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err = ensureVideoTx(ctx, tx, videoID); err != nil {
		return nil, err
	}
	for _, input := range inputs {
		var captionID int64
		err = tx.QueryRow(ctx, `INSERT INTO video_captions(video_id,sentence_order,start_time_ms,end_time_ms,content,pinyin_or_furigana)
			VALUES($1,$2,$3,$4,$5,NULLIF($6,'')) RETURNING id`, videoID, input.SentenceOrder, input.StartTimeMS,
			input.EndTimeMS, input.Content, input.PinyinOrFurigana).Scan(&captionID)
		if err != nil {
			return nil, mapWriteError("ImportCaptions", err)
		}
		if err = replaceTranslations(ctx, tx, captionID, input.Translations); err != nil {
			return nil, mapWriteError("ImportCaptions translations", err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("CaptionRepo - ImportCaptions - commit: %w", err)
	}
	result, err := r.ListCaptions(ctx, videoID, entity.CaptionFilter{Limit: maxInt, Offset: 0})
	return result.Items, err
}

const maxInt = int(^uint(0) >> 1)

func (r *Repo) UpdateCaption(ctx context.Context, videoID, captionID int64, input entity.CaptionInput) (entity.Caption, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return entity.Caption{}, fmt.Errorf("CaptionRepo - UpdateCaption - begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	result, err := tx.Exec(ctx, `UPDATE video_captions SET sentence_order=$3,start_time_ms=$4,end_time_ms=$5,content=$6,
		pinyin_or_furigana=NULLIF($7,'') WHERE id=$1 AND video_id=$2`, captionID, videoID, input.SentenceOrder,
		input.StartTimeMS, input.EndTimeMS, input.Content, input.PinyinOrFurigana)
	if err != nil {
		return entity.Caption{}, mapWriteError("UpdateCaption", err)
	}
	if result.RowsAffected() == 0 {
		return entity.Caption{}, entity.ErrCaptionNotFound
	}
	if _, err = tx.Exec(ctx, `DELETE FROM caption_translations WHERE caption_id=$1`, captionID); err != nil {
		return entity.Caption{}, fmt.Errorf("CaptionRepo - UpdateCaption - delete translations: %w", err)
	}
	if err = replaceTranslations(ctx, tx, captionID, input.Translations); err != nil {
		return entity.Caption{}, mapWriteError("UpdateCaption translations", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return entity.Caption{}, fmt.Errorf("CaptionRepo - UpdateCaption - commit: %w", err)
	}
	return r.getCaption(ctx, videoID, captionID)
}

func (r *Repo) DeleteCaption(ctx context.Context, videoID, captionID int64) error {
	result, err := r.Pool.Exec(ctx, `DELETE FROM video_captions WHERE id=$1 AND video_id=$2`, captionID, videoID)
	if err != nil {
		return fmt.Errorf("CaptionRepo - DeleteCaption: %w", err)
	}
	if result.RowsAffected() == 0 {
		return entity.ErrCaptionNotFound
	}
	return nil
}

func (r *Repo) getCaption(ctx context.Context, videoID, captionID int64) (entity.Caption, error) {
	item, err := scanCaption(r.Pool.QueryRow(ctx, `SELECT `+captionColumns+` FROM video_captions c WHERE c.video_id=$1 AND c.id=$2`, videoID, captionID))
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.Caption{}, entity.ErrCaptionNotFound
	}
	return item, err
}

func (r *Repo) ensureVideo(ctx context.Context, videoID int64) error {
	var exists bool
	if err := r.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM videos WHERE id=$1)`, videoID).Scan(&exists); err != nil {
		return fmt.Errorf("CaptionRepo - ensureVideo: %w", err)
	}
	if !exists {
		return entity.ErrVideoNotFound
	}
	return nil
}

func ensureVideoTx(ctx context.Context, tx pgx.Tx, videoID int64) error {
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM videos WHERE id=$1)`, videoID).Scan(&exists); err != nil {
		return fmt.Errorf("CaptionRepo - ensureVideoTx: %w", err)
	}
	if !exists {
		return entity.ErrVideoNotFound
	}
	return nil
}

type scanner interface{ Scan(...any) error }

func scanCaption(row scanner) (entity.Caption, error) {
	var item entity.Caption
	var translationsJSON []byte
	err := row.Scan(&item.ID, &item.VideoID, &item.SentenceOrder, &item.StartTimeMS, &item.EndTimeMS,
		&item.Content, &item.PinyinOrFurigana, &item.CreatedAt, &translationsJSON)
	if err != nil {
		return item, err
	}
	if err = json.Unmarshal(translationsJSON, &item.Translations); err != nil {
		return item, fmt.Errorf("decode translations: %w", err)
	}
	return item, nil
}

func replaceTranslations(ctx context.Context, tx pgx.Tx, captionID int64, translations []entity.CaptionTranslationInput) error {
	for _, translation := range translations {
		if _, err := tx.Exec(ctx, `INSERT INTO caption_translations(caption_id,language_id,translated_text) VALUES($1,$2,$3)`,
			captionID, translation.LanguageID, translation.Text); err != nil {
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
			return entity.ErrCaptionExists
		case "23503":
			return entity.ErrInvalidReference
		case "23514":
			return entity.ErrInvalidCaption
		}
	}
	return fmt.Errorf("CaptionRepo - %s: %w", operation, err)
}
