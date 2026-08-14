package learningcontent

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

func New(pg *postgres.Postgres) repo.LearningContentRepo { return &Repo{Postgres: pg} }

func (r *Repo) GetQuizzes(ctx context.Context, videoID int64) ([]entity.LearningQuiz, error) {
	rows, err := r.Pool.Query(ctx, `SELECT q.id,q.display_order,q.question,q.correct_option,q.explanation,
		o.display_order,o.content FROM video_quizzes q JOIN video_quiz_options o ON o.quiz_id=q.id
		WHERE q.video_id=$1 ORDER BY q.display_order,o.display_order`, videoID)
	if err != nil {
		return nil, fmt.Errorf("LearningContentRepo - GetQuizzes: %w", err)
	}
	defer rows.Close()
	items := make([]entity.LearningQuiz, 0)
	positions := make(map[int64]int)
	for rows.Next() {
		var quizID int64
		var quizOrder, correctOption, optionOrder int
		var question, explanation, option string
		if err = rows.Scan(&quizID, &quizOrder, &question, &correctOption, &explanation, &optionOrder, &option); err != nil {
			return nil, fmt.Errorf("LearningContentRepo - GetQuizzes - scan: %w", err)
		}
		position, found := positions[quizID]
		if !found {
			position = len(items)
			positions[quizID] = position
			items = append(items, entity.LearningQuiz{ID: quizID, Order: quizOrder, Question: question,
				CorrectOption: correctOption, Explanation: explanation, Options: []entity.LearningQuizOption{}})
		}
		items[position].Options = append(items[position].Options, entity.LearningQuizOption{Order: optionOrder, Content: option})
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("LearningContentRepo - GetQuizzes - rows: %w", err)
	}
	if len(items) == 0 {
		return nil, entity.ErrLearningContentNotFound
	}
	return items, nil
}

func (r *Repo) SaveQuizzesIfAbsent(ctx context.Context, videoID int64, quizzes []entity.GeneratedQuiz) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("LearningContentRepo - SaveQuizzes - begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, videoID); err != nil {
		return fmt.Errorf("LearningContentRepo - SaveQuizzes - lock: %w", err)
	}
	var exists bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM video_quizzes WHERE video_id=$1)`, videoID).Scan(&exists); err != nil {
		return fmt.Errorf("LearningContentRepo - SaveQuizzes - check: %w", err)
	}
	if !exists {
		for index, quiz := range quizzes {
			var quizID int64
			if err = tx.QueryRow(ctx, `INSERT INTO video_quizzes(video_id,display_order,question,correct_option,explanation)
				VALUES($1,$2,$3,$4,$5) RETURNING id`, videoID, index+1, quiz.Question, quiz.CorrectOption, quiz.Explanation).Scan(&quizID); err != nil {
				return fmt.Errorf("LearningContentRepo - SaveQuizzes - insert: %w", err)
			}
			for optionIndex, option := range quiz.Options {
				if _, err = tx.Exec(ctx, `INSERT INTO video_quiz_options(quiz_id,display_order,content) VALUES($1,$2,$3)`, quizID, optionIndex, option); err != nil {
					return fmt.Errorf("LearningContentRepo - SaveQuizzes - option: %w", err)
				}
			}
		}
	}
	return tx.Commit(ctx)
}

const dictionaryColumns = `d.id,d.word,d.source_language_id,d.target_language_id,COALESCE(d.phonetic_or_pinyin,''),
	COALESCE(d.part_of_speech,''),d.meaning,COALESCE(d.example_1_sentence,''),COALESCE(d.example_1_translation,''),
	COALESCE(d.example_2_sentence,''),COALESCE(d.example_2_translation,''),d.created_at`

func (r *Repo) GetLocalizedLearningContent(ctx context.Context, videoID int64, languageID int) (entity.VideoSummary, []entity.DictionaryEntry, error) {
	var summary entity.VideoSummary
	err := r.Pool.QueryRow(ctx, `SELECT s.video_id,s.language_id,l.code,s.content,s.created_at
		FROM video_summaries s JOIN languages l ON l.id=s.language_id WHERE s.video_id=$1 AND s.language_id=$2`, videoID, languageID).
		Scan(&summary.VideoID, &summary.LanguageID, &summary.LanguageCode, &summary.Content, &summary.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return summary, nil, entity.ErrLearningContentNotFound
	}
	if err != nil {
		return summary, nil, fmt.Errorf("LearningContentRepo - GetLocalized - summary: %w", err)
	}
	rows, err := r.Pool.Query(ctx, `SELECT `+dictionaryColumns+` FROM video_dictionary_entries vde
		JOIN dictionary d ON d.id=vde.dictionary_id WHERE vde.video_id=$1 AND vde.target_language_id=$2 ORDER BY vde.display_order`, videoID, languageID)
	if err != nil {
		return summary, nil, fmt.Errorf("LearningContentRepo - GetLocalized - vocabulary: %w", err)
	}
	defer rows.Close()
	items := make([]entity.DictionaryEntry, 0)
	for rows.Next() {
		var item entity.DictionaryEntry
		if err = rows.Scan(&item.ID, &item.Word, &item.SourceLanguageID, &item.TargetLanguageID, &item.PhoneticOrPinyin,
			&item.PartOfSpeech, &item.Meaning, &item.Example1Sentence, &item.Example1Translation,
			&item.Example2Sentence, &item.Example2Translation, &item.CreatedAt); err != nil {
			return summary, nil, fmt.Errorf("LearningContentRepo - GetLocalized - scan: %w", err)
		}
		items = append(items, item)
	}
	return summary, items, rows.Err()
}

func (r *Repo) SaveLocalizedLearningContentIfAbsent(ctx context.Context, summary entity.VideoSummary, dictionaryIDs []int64) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("LearningContentRepo - SaveLocalized - begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	lockKey := summary.VideoID*100000 + int64(summary.LanguageID)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, lockKey); err != nil {
		return fmt.Errorf("LearningContentRepo - SaveLocalized - lock: %w", err)
	}
	result, err := tx.Exec(ctx, `INSERT INTO video_summaries(video_id,language_id,content) VALUES($1,$2,$3)
		ON CONFLICT(video_id,language_id) DO NOTHING`, summary.VideoID, summary.LanguageID, summary.Content)
	if err != nil {
		return fmt.Errorf("LearningContentRepo - SaveLocalized - summary: %w", err)
	}
	if result.RowsAffected() > 0 {
		for index, dictionaryID := range dictionaryIDs {
			if _, err = tx.Exec(ctx, `INSERT INTO video_dictionary_entries(video_id,target_language_id,dictionary_id,display_order)
				VALUES($1,$2,$3,$4)`, summary.VideoID, summary.LanguageID, dictionaryID, index+1); err != nil {
				return fmt.Errorf("LearningContentRepo - SaveLocalized - vocabulary: %w", err)
			}
		}
	}
	return tx.Commit(ctx)
}
