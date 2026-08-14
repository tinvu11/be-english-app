package dictionary

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

func New(pg *postgres.Postgres) repo.DictionaryRepo { return &Repo{Postgres: pg} }

const columns = `id,word,source_language_id,target_language_id,COALESCE(phonetic_or_pinyin,''),
	COALESCE(part_of_speech,''),COALESCE(audio_url,''),meaning,COALESCE(example_1_sentence,''),COALESCE(example_1_translation,''),
	COALESCE(example_2_sentence,''),COALESCE(example_2_translation,''),created_at`

func (r *Repo) FindDictionaryEntry(ctx context.Context, word string, sourceLanguageID, targetLanguageID int) (entity.DictionaryEntry, error) {
	entry, err := scan(r.Pool.QueryRow(ctx, `SELECT `+columns+` FROM dictionary
		WHERE word=$1 AND source_language_id=$2 AND target_language_id=$3`, word, sourceLanguageID, targetLanguageID))
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.DictionaryEntry{}, entity.ErrDictionaryNotFound
	}
	if err != nil {
		return entity.DictionaryEntry{}, fmt.Errorf("DictionaryRepo - FindDictionaryEntry: %w", err)
	}
	return entry, nil
}

func (r *Repo) UpsertDictionaryEntry(ctx context.Context, input entity.DictionaryInput) (entity.DictionaryEntry, bool, error) {
	entry, err := scan(r.Pool.QueryRow(ctx, `INSERT INTO dictionary(word,source_language_id,target_language_id,
		phonetic_or_pinyin,part_of_speech,audio_url,meaning,example_1_sentence,example_1_translation,example_2_sentence,example_2_translation)
		VALUES($1,$2,$3,NULLIF($4,''),NULLIF($5,''),NULLIF($6,''),$7,NULLIF($8,''),NULLIF($9,''),NULLIF($10,''),NULLIF($11,''))
		ON CONFLICT (word,source_language_id,target_language_id) DO NOTHING RETURNING `+columns,
		input.Word, input.SourceLanguageID, input.TargetLanguageID, input.PhoneticOrPinyin, input.PartOfSpeech, input.AudioURL,
		input.Meaning, input.Example1Sentence, input.Example1Translation, input.Example2Sentence, input.Example2Translation))
	if err == nil {
		return entry, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return entity.DictionaryEntry{}, false, fmt.Errorf("DictionaryRepo - UpsertDictionaryEntry: %w", err)
	}
	entry, err = r.FindDictionaryEntry(ctx, input.Word, input.SourceLanguageID, input.TargetLanguageID)
	return entry, false, err
}

func (r *Repo) GetVocabularyOverview(ctx context.Context, userID string, languages entity.VocabularyLanguages) (entity.VocabularyOverview, error) {
	rows, err := r.Pool.Query(ctx, `WITH scoped_words AS (
		SELECT uv.vocab_set_id, uv.is_learned
		FROM user_vocabularies uv
		JOIN dictionary d ON d.id=uv.dictionary_id
		WHERE uv.user_id=$1 AND d.source_language_id=$2 AND d.target_language_id=$3
	), totals AS (
		SELECT COUNT(*)::int AS total_words,
			COUNT(*) FILTER (WHERE is_learned)::int AS learned_words
		FROM scoped_words
	), categories AS (
		SELECT s.id,s.title,s.color_hex,COUNT(w.vocab_set_id)::int AS total_words,
			COUNT(w.vocab_set_id) FILTER (WHERE w.is_learned)::int AS learned_words
		FROM vocab_sets s
		LEFT JOIN scoped_words w ON w.vocab_set_id=s.id
		WHERE s.user_id=$1 AND s.source_language_id=$2 AND s.target_language_id=$3
		GROUP BY s.id,s.title,s.color_hex,s.created_at
		ORDER BY s.created_at DESC,s.id DESC
	)
	SELECT t.total_words,t.learned_words,c.id,c.title,c.color_hex,c.total_words,c.learned_words
	FROM totals t LEFT JOIN categories c ON TRUE
	ORDER BY c.id DESC`, userID, languages.SourceLanguageID, languages.TargetLanguageID)
	if err != nil {
		return entity.VocabularyOverview{}, fmt.Errorf("DictionaryRepo - GetVocabularyOverview: %w", err)
	}
	defer rows.Close()

	overview := entity.VocabularyOverview{Categories: make([]entity.VocabularyCategorySummary, 0)}
	for rows.Next() {
		var categoryID *int64
		var title *string
		var colorHex *string
		var categoryTotal, categoryLearned *int
		if err = rows.Scan(&overview.TotalWords, &overview.LearnedWords, &categoryID, &title, &colorHex, &categoryTotal, &categoryLearned); err != nil {
			return entity.VocabularyOverview{}, fmt.Errorf("DictionaryRepo - GetVocabularyOverview - scan: %w", err)
		}
		if categoryID != nil {
			category := entity.VocabularyCategorySummary{ID: *categoryID, Title: *title, ColorHex: *colorHex, TotalWords: *categoryTotal, LearnedWords: *categoryLearned}
			category.UnlearnedWords = category.TotalWords - category.LearnedWords
			overview.Categories = append(overview.Categories, category)
		}
	}
	if err = rows.Err(); err != nil {
		return entity.VocabularyOverview{}, fmt.Errorf("DictionaryRepo - GetVocabularyOverview - rows: %w", err)
	}
	overview.UnlearnedWords = overview.TotalWords - overview.LearnedWords
	return overview, nil
}

type scanner interface{ Scan(...any) error }

func scan(row scanner) (entity.DictionaryEntry, error) {
	var entry entity.DictionaryEntry
	err := row.Scan(&entry.ID, &entry.Word, &entry.SourceLanguageID, &entry.TargetLanguageID,
		&entry.PhoneticOrPinyin, &entry.PartOfSpeech, &entry.AudioURL, &entry.Meaning, &entry.Example1Sentence,
		&entry.Example1Translation, &entry.Example2Sentence, &entry.Example2Translation, &entry.CreatedAt)
	return entry, err
}

func (r *Repo) CreateVocabularySet(ctx context.Context, userID, title, colorHex string, languages entity.VocabularyLanguages) (entity.VocabularySet, error) {
	var item entity.VocabularySet
	err := r.Pool.QueryRow(ctx, `INSERT INTO vocab_sets(user_id,title,color_hex,source_language_id,target_language_id) VALUES($1,$2,$3,$4,$5)
		RETURNING id,title,color_hex,source_language_id,target_language_id,0,created_at,updated_at`, userID, title, colorHex,
		languages.SourceLanguageID, languages.TargetLanguageID).
		Scan(&item.ID, &item.Title, &item.ColorHex, &item.SourceLanguageID, &item.TargetLanguageID, &item.WordCount, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return item, fmt.Errorf("DictionaryRepo - CreateVocabularySet: %w", err)
	}
	return item, nil
}

func (r *Repo) ListVocabularySets(ctx context.Context, userID string, languages entity.VocabularyLanguages) ([]entity.VocabularySet, error) {
	rows, err := r.Pool.Query(ctx, `SELECT sets.id,sets.title,sets.color_hex,sets.source_language_id,sets.target_language_id,COUNT(vocab.id),sets.created_at,sets.updated_at
		FROM vocab_sets sets LEFT JOIN user_vocabularies vocab ON vocab.vocab_set_id=sets.id AND vocab.user_id=sets.user_id
		WHERE sets.user_id=$1 AND sets.source_language_id=$2 AND sets.target_language_id=$3
		GROUP BY sets.id ORDER BY sets.created_at DESC,sets.id DESC`, userID, languages.SourceLanguageID, languages.TargetLanguageID)
	if err != nil {
		return nil, fmt.Errorf("DictionaryRepo - ListVocabularySets: %w", err)
	}
	defer rows.Close()
	items := make([]entity.VocabularySet, 0)
	for rows.Next() {
		var item entity.VocabularySet
		if err = rows.Scan(&item.ID, &item.Title, &item.ColorHex, &item.SourceLanguageID, &item.TargetLanguageID, &item.WordCount, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("DictionaryRepo - ListVocabularySets - scan: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repo) UpdateVocabularySet(ctx context.Context, userID string, id int64, title, colorHex string, languages entity.VocabularyLanguages) (entity.VocabularySet, error) {
	var item entity.VocabularySet
	err := r.Pool.QueryRow(ctx, `UPDATE vocab_sets SET title=$3,color_hex=$4,updated_at=CURRENT_TIMESTAMP
		WHERE id=$1 AND user_id=$2 AND source_language_id=$5 AND target_language_id=$6 RETURNING id,title,color_hex,source_language_id,target_language_id,
		(SELECT COUNT(*) FROM user_vocabularies WHERE vocab_set_id=$1 AND user_id=$2),created_at,updated_at`, id, userID, title, colorHex, languages.SourceLanguageID, languages.TargetLanguageID).
		Scan(&item.ID, &item.Title, &item.ColorHex, &item.SourceLanguageID, &item.TargetLanguageID, &item.WordCount, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return item, entity.ErrVocabularySetNotFound
	}
	if err != nil {
		return item, fmt.Errorf("DictionaryRepo - UpdateVocabularySet: %w", err)
	}
	return item, nil
}

func (r *Repo) DeleteVocabularySet(ctx context.Context, userID string, id int64, languages entity.VocabularyLanguages) error {
	result, err := r.Pool.Exec(ctx, `DELETE FROM vocab_sets WHERE id=$1 AND user_id=$2 AND source_language_id=$3 AND target_language_id=$4`, id, userID, languages.SourceLanguageID, languages.TargetLanguageID)
	if err != nil {
		return fmt.Errorf("DictionaryRepo - DeleteVocabularySet: %w", err)
	}
	if result.RowsAffected() == 0 {
		return entity.ErrVocabularySetNotFound
	}
	return nil
}

const userVocabularyColumns = `uv.id,uv.vocab_set_id,uv.caption_id,uv.is_learned,uv.learned_at,uv.created_at,uv.updated_at,
	d.id,d.word,d.source_language_id,d.target_language_id,COALESCE(d.phonetic_or_pinyin,''),COALESCE(d.part_of_speech,''),COALESCE(d.audio_url,''),
	d.meaning,COALESCE(d.example_1_sentence,''),COALESCE(d.example_1_translation,''),
	COALESCE(d.example_2_sentence,''),COALESCE(d.example_2_translation,''),d.created_at`

func (r *Repo) CreateUserVocabulary(ctx context.Context, userID string, input entity.UserVocabularyInput, languages entity.VocabularyLanguages) (entity.UserVocabulary, error) {
	var id int64
	err := r.Pool.QueryRow(ctx, `INSERT INTO user_vocabularies(user_id,vocab_set_id,dictionary_id,caption_id)
		SELECT $1,$2,$3,$4 WHERE EXISTS (SELECT 1 FROM dictionary d WHERE d.id=$3 AND d.source_language_id=$5 AND d.target_language_id=$6)
		AND EXISTS (SELECT 1 FROM vocab_sets s WHERE s.id=$2 AND s.user_id=$1 AND s.source_language_id=$5 AND s.target_language_id=$6)
		RETURNING id`, userID, input.VocabSetID, input.DictionaryID, input.CaptionID, languages.SourceLanguageID, languages.TargetLanguageID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.UserVocabulary{}, entity.ErrVocabularyLanguageMismatch
	}
	if err != nil {
		return entity.UserVocabulary{}, mapUserVocabularyError(err)
	}
	return r.getUserVocabulary(ctx, userID, id)
}

func (r *Repo) ListUserVocabularies(ctx context.Context, userID string, filter entity.UserVocabularyFilter, languages entity.VocabularyLanguages) ([]entity.UserVocabulary, error) {
	rows, err := r.Pool.Query(ctx, `SELECT `+userVocabularyColumns+` FROM user_vocabularies uv
		JOIN dictionary d ON d.id=uv.dictionary_id WHERE uv.user_id=$1
		AND ($2::bigint IS NULL OR uv.vocab_set_id=$2) AND ($3='' OR strpos(lower(d.word),lower($3))>0)
		AND d.source_language_id=$4 AND d.target_language_id=$5
		ORDER BY uv.created_at DESC,uv.id DESC`, userID, filter.VocabSetID, filter.Search, languages.SourceLanguageID, languages.TargetLanguageID)
	if err != nil {
		return nil, fmt.Errorf("DictionaryRepo - ListUserVocabularies: %w", err)
	}
	defer rows.Close()
	items := make([]entity.UserVocabulary, 0)
	for rows.Next() {
		item, scanErr := scanUserVocabulary(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("DictionaryRepo - ListUserVocabularies - scan: %w", scanErr)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repo) ListUnlearnedVocabularies(ctx context.Context, userID string, filter entity.UnlearnedVocabularyFilter, languages entity.VocabularyLanguages) ([]entity.UserVocabulary, error) {
	query := `SELECT ` + userVocabularyColumns + ` FROM user_vocabularies uv
		JOIN dictionary d ON d.id=uv.dictionary_id WHERE uv.user_id=$1 AND uv.is_learned=FALSE
		AND ($2::bigint IS NULL OR uv.vocab_set_id=$2) AND d.source_language_id=$3 AND d.target_language_id=$4`
	args := []any{userID, filter.VocabSetID, languages.SourceLanguageID, languages.TargetLanguageID}
	if filter.Limit == nil {
		query += ` ORDER BY uv.created_at DESC,uv.id DESC`
	} else {
		query += ` ORDER BY random() LIMIT $5`
		args = append(args, *filter.Limit)
	}
	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("DictionaryRepo - ListUnlearnedVocabularies: %w", err)
	}
	defer rows.Close()
	items := make([]entity.UserVocabulary, 0)
	for rows.Next() {
		item, scanErr := scanUserVocabulary(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("DictionaryRepo - ListUnlearnedVocabularies - scan: %w", scanErr)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repo) UpdateUserVocabulary(ctx context.Context, userID string, id int64, input entity.UserVocabularyUpdate, languages entity.VocabularyLanguages) (entity.UserVocabulary, error) {
	result, err := r.Pool.Exec(ctx, `UPDATE user_vocabularies uv SET vocab_set_id=$3,caption_id=$4,is_learned=$5,
		learned_at=CASE WHEN $5 THEN COALESCE(learned_at,CURRENT_TIMESTAMP) ELSE NULL END,updated_at=CURRENT_TIMESTAMP
		WHERE uv.id=$1 AND uv.user_id=$2 AND EXISTS (SELECT 1 FROM dictionary d WHERE d.id=uv.dictionary_id AND d.source_language_id=$6 AND d.target_language_id=$7)
		AND EXISTS (SELECT 1 FROM vocab_sets s WHERE s.id=$3 AND s.user_id=$2 AND s.source_language_id=$6 AND s.target_language_id=$7)`, id, userID, input.VocabSetID, input.CaptionID, input.IsLearned, languages.SourceLanguageID, languages.TargetLanguageID)
	if err != nil {
		return entity.UserVocabulary{}, mapUserVocabularyError(err)
	}
	if result.RowsAffected() == 0 {
		return entity.UserVocabulary{}, entity.ErrUserVocabularyNotFound
	}
	return r.getUserVocabulary(ctx, userID, id)
}

func (r *Repo) DeleteUserVocabulary(ctx context.Context, userID string, id int64, languages entity.VocabularyLanguages) error {
	result, err := r.Pool.Exec(ctx, `DELETE FROM user_vocabularies uv USING dictionary d WHERE uv.id=$1 AND uv.user_id=$2 AND d.id=uv.dictionary_id AND d.source_language_id=$3 AND d.target_language_id=$4`, id, userID, languages.SourceLanguageID, languages.TargetLanguageID)
	if err != nil {
		return fmt.Errorf("DictionaryRepo - DeleteUserVocabulary: %w", err)
	}
	if result.RowsAffected() == 0 {
		return entity.ErrUserVocabularyNotFound
	}
	return nil
}

func (r *Repo) getUserVocabulary(ctx context.Context, userID string, id int64) (entity.UserVocabulary, error) {
	item, err := scanUserVocabulary(r.Pool.QueryRow(ctx, `SELECT `+userVocabularyColumns+` FROM user_vocabularies uv
		JOIN dictionary d ON d.id=uv.dictionary_id WHERE uv.id=$1 AND uv.user_id=$2`, id, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return item, entity.ErrUserVocabularyNotFound
	}
	return item, err
}

func scanUserVocabulary(row scanner) (entity.UserVocabulary, error) {
	var item entity.UserVocabulary
	err := row.Scan(&item.ID, &item.VocabSetID, &item.CaptionID, &item.IsLearned, &item.LearnedAt,
		&item.CreatedAt, &item.UpdatedAt, &item.Entry.ID, &item.Entry.Word, &item.Entry.SourceLanguageID,
		&item.Entry.TargetLanguageID, &item.Entry.PhoneticOrPinyin, &item.Entry.PartOfSpeech, &item.Entry.AudioURL, &item.Entry.Meaning,
		&item.Entry.Example1Sentence, &item.Entry.Example1Translation, &item.Entry.Example2Sentence,
		&item.Entry.Example2Translation, &item.Entry.CreatedAt)
	return item, err
}

func mapUserVocabularyError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return entity.ErrUserVocabularyExists
		case "23503":
			return entity.ErrInvalidReference
		}
	}
	return fmt.Errorf("DictionaryRepo - user vocabulary write: %w", err)
}
