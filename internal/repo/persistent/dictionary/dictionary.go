package dictionary

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

func New(pg *postgres.Postgres) repo.DictionaryRepo { return &Repo{Postgres: pg} }

const columns = `id,word,source_language_id,target_language_id,COALESCE(phonetic_or_pinyin,''),
	COALESCE(part_of_speech,''),meaning,COALESCE(example_1_sentence,''),COALESCE(example_1_translation,''),
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
		phonetic_or_pinyin,part_of_speech,meaning,example_1_sentence,example_1_translation,example_2_sentence,example_2_translation)
		VALUES($1,$2,$3,NULLIF($4,''),NULLIF($5,''),$6,NULLIF($7,''),NULLIF($8,''),NULLIF($9,''),NULLIF($10,''))
		ON CONFLICT (word,source_language_id,target_language_id) DO NOTHING RETURNING `+columns,
		input.Word, input.SourceLanguageID, input.TargetLanguageID, input.PhoneticOrPinyin, input.PartOfSpeech,
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

type scanner interface{ Scan(...any) error }

func scan(row scanner) (entity.DictionaryEntry, error) {
	var entry entity.DictionaryEntry
	err := row.Scan(&entry.ID, &entry.Word, &entry.SourceLanguageID, &entry.TargetLanguageID,
		&entry.PhoneticOrPinyin, &entry.PartOfSpeech, &entry.Meaning, &entry.Example1Sentence,
		&entry.Example1Translation, &entry.Example2Sentence, &entry.Example2Translation, &entry.CreatedAt)
	return entry, err
}
