package shadowing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/jackc/pgx/v5"
)

type Repo struct{ *postgres.Postgres }

func New(pg *postgres.Postgres) repo.ShadowingRepo { return &Repo{Postgres: pg} }

func (r *Repo) GetPrompt(ctx context.Context, videoID, captionID int64) (entity.ShadowingPrompt, error) {
	var prompt entity.ShadowingPrompt
	err := r.Pool.QueryRow(ctx, `SELECT c.video_id,c.id,c.content,l.code
		FROM video_captions c JOIN videos v ON v.id=c.video_id JOIN languages l ON l.id=v.language_id
		WHERE c.video_id=$1 AND c.id=$2`, videoID, captionID).
		Scan(&prompt.VideoID, &prompt.CaptionID, &prompt.ReferenceText, &prompt.LanguageCode)
	if errors.Is(err, pgx.ErrNoRows) {
		return prompt, entity.ErrCaptionNotFound
	}
	if err != nil {
		return prompt, fmt.Errorf("ShadowingRepo - GetPrompt: %w", err)
	}
	return prompt, nil
}

func (r *Repo) SaveAttempt(ctx context.Context, userID string, prompt entity.ShadowingPrompt, assessment entity.PronunciationAssessment) (entity.ShadowingAttempt, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return entity.ShadowingAttempt{}, fmt.Errorf("ShadowingRepo - SaveAttempt - begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var attemptID int64
	var createdAt entity.ShadowingAttempt
	err = tx.QueryRow(ctx, `INSERT INTO shadowing_attempts(user_id,caption_id,accuracy_score,fluency_score,
		completeness_score,pronunciation_score,provider_response_json,provider,prosody_score,locale,duration_ms,reference_text)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT(user_id,caption_id) DO UPDATE SET
			accuracy_score=EXCLUDED.accuracy_score,
			fluency_score=EXCLUDED.fluency_score,
			completeness_score=EXCLUDED.completeness_score,
			pronunciation_score=EXCLUDED.pronunciation_score,
			provider_response_json=EXCLUDED.provider_response_json,
			provider=EXCLUDED.provider,
			prosody_score=EXCLUDED.prosody_score,
			locale=EXCLUDED.locale,
			duration_ms=EXCLUDED.duration_ms,
			reference_text=EXCLUDED.reference_text,
			created_at=CURRENT_TIMESTAMP
		WHERE shadowing_attempts.pronunciation_score IS NULL
			OR EXCLUDED.pronunciation_score > shadowing_attempts.pronunciation_score
		RETURNING id,created_at`, userID, prompt.CaptionID, assessment.AccuracyScore, assessment.FluencyScore,
		assessment.CompletenessScore, assessment.PronunciationScore, assessment.ProviderResponse, assessment.Provider,
		assessment.ProsodyScore, prompt.LanguageCode, assessment.DurationMS, prompt.ReferenceText).
		Scan(&attemptID, &createdAt.CreatedAt)
	improved := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return entity.ShadowingAttempt{}, fmt.Errorf("ShadowingRepo - SaveAttempt - insert: %w", err)
	}
	if improved {
		if _, err = tx.Exec(ctx, `DELETE FROM shadowing_word_results WHERE attempt_id=$1`, attemptID); err != nil {
			return entity.ShadowingAttempt{}, fmt.Errorf("ShadowingRepo - SaveAttempt - clear words: %w", err)
		}
	}
	for _, word := range assessment.Words {
		if !improved {
			break
		}
		var phonemes any
		if len(word.Phonemes) > 0 {
			phonemes = word.Phonemes
		}
		if _, err = tx.Exec(ctx, `INSERT INTO shadowing_word_results(attempt_id,word_order,word,accuracy_score,error_type,phonemes_json)
			VALUES($1,$2,$3,$4,NULLIF($5,''),$6)`, attemptID, word.Order, word.Word, word.AccuracyScore, word.ErrorType, phonemes); err != nil {
			return entity.ShadowingAttempt{}, fmt.Errorf("ShadowingRepo - SaveAttempt - word: %w", err)
		}
	}
	if !improved {
		attempt, loadErr := loadBestAttempt(ctx, tx, userID, prompt)
		if loadErr != nil {
			return entity.ShadowingAttempt{}, loadErr
		}
		if err = tx.Commit(ctx); err != nil {
			return entity.ShadowingAttempt{}, fmt.Errorf("ShadowingRepo - SaveAttempt - commit: %w", err)
		}
		return attempt, nil
	}
	if err = tx.Commit(ctx); err != nil {
		return entity.ShadowingAttempt{}, fmt.Errorf("ShadowingRepo - SaveAttempt - commit: %w", err)
	}
	attempt := createdAt
	attempt.ID = attemptID
	attempt.VideoID, attempt.CaptionID = prompt.VideoID, prompt.CaptionID
	attempt.Provider, attempt.Locale, attempt.ReferenceText = assessment.Provider, prompt.LanguageCode, prompt.ReferenceText
	attempt.AccuracyScore, attempt.FluencyScore = assessment.AccuracyScore, assessment.FluencyScore
	attempt.CompletenessScore, attempt.PronunciationScore = assessment.CompletenessScore, assessment.PronunciationScore
	attempt.ProsodyScore, attempt.DurationMS, attempt.Words = assessment.ProsodyScore, assessment.DurationMS, assessment.Words
	return attempt, nil
}

func loadBestAttempt(ctx context.Context, tx pgx.Tx, userID string, prompt entity.ShadowingPrompt) (entity.ShadowingAttempt, error) {
	var attempt entity.ShadowingAttempt
	err := tx.QueryRow(ctx, `SELECT id,provider,COALESCE(locale,''),reference_text,COALESCE(accuracy_score,0),
		COALESCE(fluency_score,0),COALESCE(completeness_score,0),COALESCE(pronunciation_score,0),prosody_score,
		COALESCE(duration_ms,0),created_at FROM shadowing_attempts WHERE user_id=$1 AND caption_id=$2`, userID, prompt.CaptionID).
		Scan(&attempt.ID, &attempt.Provider, &attempt.Locale, &attempt.ReferenceText, &attempt.AccuracyScore,
			&attempt.FluencyScore, &attempt.CompletenessScore, &attempt.PronunciationScore, &attempt.ProsodyScore,
			&attempt.DurationMS, &attempt.CreatedAt)
	if err != nil {
		return attempt, fmt.Errorf("ShadowingRepo - SaveAttempt - load best: %w", err)
	}
	attempt.VideoID, attempt.CaptionID = prompt.VideoID, prompt.CaptionID
	attempt.Words = make([]entity.ShadowingWordResult, 0)
	rows, err := tx.Query(ctx, `SELECT word_order,word,COALESCE(accuracy_score,0),COALESCE(error_type,''),phonemes_json
		FROM shadowing_word_results WHERE attempt_id=$1 ORDER BY word_order`, attempt.ID)
	if err != nil {
		return attempt, fmt.Errorf("ShadowingRepo - SaveAttempt - load words: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var word entity.ShadowingWordResult
		var phonemes []byte
		if err = rows.Scan(&word.Order, &word.Word, &word.AccuracyScore, &word.ErrorType, &phonemes); err != nil {
			return attempt, fmt.Errorf("ShadowingRepo - SaveAttempt - load word: %w", err)
		}
		if json.Valid(phonemes) {
			word.Phonemes = phonemes
		}
		attempt.Words = append(attempt.Words, word)
	}
	return attempt, rows.Err()
}

func (r *Repo) ListAttempts(ctx context.Context, userID string, videoID, captionID int64) (entity.ShadowingAttemptList, error) {
	rows, err := r.Pool.Query(ctx, `SELECT a.id,c.video_id,a.caption_id,a.provider,COALESCE(a.locale,''),a.reference_text,
		COALESCE(a.accuracy_score,0),COALESCE(a.fluency_score,0),COALESCE(a.completeness_score,0),
		COALESCE(a.pronunciation_score,0),a.prosody_score,COALESCE(a.duration_ms,0),a.created_at,COUNT(*) OVER()
		FROM shadowing_attempts a JOIN video_captions c ON c.id=a.caption_id
		WHERE a.user_id=$1 AND c.video_id=$2 AND ($3=0 OR a.caption_id=$3)
		ORDER BY a.created_at DESC,a.id DESC`, userID, videoID, captionID)
	if err != nil {
		return entity.ShadowingAttemptList{}, fmt.Errorf("ShadowingRepo - ListAttempts: %w", err)
	}
	defer rows.Close()
	result := entity.ShadowingAttemptList{Items: make([]entity.ShadowingAttempt, 0)}
	positions := make(map[int64]int)
	for rows.Next() {
		var item entity.ShadowingAttempt
		if err = rows.Scan(&item.ID, &item.VideoID, &item.CaptionID, &item.Provider, &item.Locale, &item.ReferenceText,
			&item.AccuracyScore, &item.FluencyScore, &item.CompletenessScore, &item.PronunciationScore,
			&item.ProsodyScore, &item.DurationMS, &item.CreatedAt, &result.Total); err != nil {
			return result, fmt.Errorf("ShadowingRepo - ListAttempts - scan: %w", err)
		}
		item.Words = make([]entity.ShadowingWordResult, 0)
		positions[item.ID] = len(result.Items)
		result.Items = append(result.Items, item)
	}
	if err = rows.Err(); err != nil || len(result.Items) == 0 {
		return result, err
	}
	ids := make([]int64, 0, len(result.Items))
	for _, item := range result.Items {
		ids = append(ids, item.ID)
	}
	wordRows, err := r.Pool.Query(ctx, `SELECT attempt_id,word_order,word,COALESCE(accuracy_score,0),COALESCE(error_type,''),phonemes_json
		FROM shadowing_word_results WHERE attempt_id=ANY($1) ORDER BY attempt_id,word_order`, ids)
	if err != nil {
		return result, fmt.Errorf("ShadowingRepo - ListAttempts - words: %w", err)
	}
	defer wordRows.Close()
	for wordRows.Next() {
		var attemptID int64
		var item entity.ShadowingWordResult
		var phonemes []byte
		if err = wordRows.Scan(&attemptID, &item.Order, &item.Word, &item.AccuracyScore, &item.ErrorType, &phonemes); err != nil {
			return result, fmt.Errorf("ShadowingRepo - ListAttempts - word scan: %w", err)
		}
		if json.Valid(phonemes) {
			item.Phonemes = phonemes
		}
		position, ok := positions[attemptID]
		if ok {
			result.Items[position].Words = append(result.Items[position].Words, item)
		}
	}
	return result, wordRows.Err()
}
