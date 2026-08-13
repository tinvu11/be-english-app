package vocabulary

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/gateway"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/internal/usecase"
)

type UseCase struct {
	dictionary repo.DictionaryRepo
	users      repo.UserRepo
	languages  repo.LanguageRepo
	translator gateway.VocabularyTranslator
}

func New(dictionary repo.DictionaryRepo, users repo.UserRepo, languages repo.LanguageRepo, translator gateway.VocabularyTranslator) usecase.Vocabulary {
	return &UseCase{dictionary: dictionary, users: users, languages: languages, translator: translator}
}

func (uc *UseCase) TranslateWord(ctx context.Context, userID, word string) (entity.VocabularyLookupResult, error) {
	word = strings.ToLower(strings.TrimSpace(word))
	if strings.TrimSpace(userID) == "" || word == "" || utf8.RuneCountInString(word) > 150 {
		return entity.VocabularyLookupResult{}, entity.ErrInvalidVocabulary
	}
	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return entity.VocabularyLookupResult{}, err
	}
	if user.TargetLanguageID == nil || *user.TargetLanguageID <= 0 {
		return entity.VocabularyLookupResult{}, entity.ErrTargetLanguageRequired
	}
	if user.NativeLanguageID == nil || *user.NativeLanguageID <= 0 {
		return entity.VocabularyLookupResult{}, entity.ErrNativeLanguageRequired
	}
	sourceID, targetID := *user.TargetLanguageID, *user.NativeLanguageID
	if sourceID == targetID {
		return entity.VocabularyLookupResult{}, entity.ErrInvalidVocabulary
	}
	entry, err := uc.dictionary.FindDictionaryEntry(ctx, word, sourceID, targetID)
	if err == nil {
		return entity.VocabularyLookupResult{Entry: entry, Reused: true}, nil
	}
	if !errors.Is(err, entity.ErrDictionaryNotFound) {
		return entity.VocabularyLookupResult{}, err
	}
	source, err := uc.languages.GetLanguage(ctx, sourceID)
	if err != nil {
		return entity.VocabularyLookupResult{}, err
	}
	target, err := uc.languages.GetLanguage(ctx, targetID)
	if err != nil {
		return entity.VocabularyLookupResult{}, err
	}
	if !source.IsActive || !target.IsActive {
		return entity.VocabularyLookupResult{}, entity.ErrInvalidLanguage
	}
	translated, err := uc.translator.TranslateVocabulary(ctx, entity.VocabularyTranslationRequest{
		Word: word, SourceLanguageCode: source.Code, SourceLanguageName: source.Name,
		TargetLanguageCode: target.Code, TargetLanguageName: target.Name,
	})
	if err != nil {
		return entity.VocabularyLookupResult{}, err
	}
	input := entity.DictionaryInput{Word: word, SourceLanguageID: sourceID, TargetLanguageID: targetID,
		PhoneticOrPinyin: strings.TrimSpace(translated.PhoneticOrPinyin), PartOfSpeech: strings.TrimSpace(translated.PartOfSpeech),
		Meaning: strings.TrimSpace(translated.Meaning), Example1Sentence: strings.TrimSpace(translated.Example1Sentence),
		Example1Translation: strings.TrimSpace(translated.Example1Translation), Example2Sentence: strings.TrimSpace(translated.Example2Sentence),
		Example2Translation: strings.TrimSpace(translated.Example2Translation)}
	if input.Meaning == "" || utf8.RuneCountInString(input.PhoneticOrPinyin) > 150 || utf8.RuneCountInString(input.PartOfSpeech) > 50 {
		return entity.VocabularyLookupResult{}, entity.ErrInvalidVocabularyTranslation
	}
	entry, created, err := uc.dictionary.UpsertDictionaryEntry(ctx, input)
	if err != nil {
		return entity.VocabularyLookupResult{}, err
	}
	return entity.VocabularyLookupResult{Entry: entry, Reused: !created}, nil
}

func (uc *UseCase) CreateSet(ctx context.Context, userID, title string) (entity.VocabularySet, error) {
	title = strings.TrimSpace(title)
	if userID == "" || title == "" || utf8.RuneCountInString(title) > 255 {
		return entity.VocabularySet{}, entity.ErrInvalidVocabularySet
	}
	return uc.dictionary.CreateVocabularySet(ctx, userID, title)
}

func (uc *UseCase) ListSets(ctx context.Context, userID string) ([]entity.VocabularySet, error) {
	if userID == "" {
		return nil, entity.ErrInvalidVocabularySet
	}
	return uc.dictionary.ListVocabularySets(ctx, userID)
}

func (uc *UseCase) UpdateSet(ctx context.Context, userID string, id int64, title string) (entity.VocabularySet, error) {
	title = strings.TrimSpace(title)
	if userID == "" || id <= 0 || title == "" || utf8.RuneCountInString(title) > 255 {
		return entity.VocabularySet{}, entity.ErrInvalidVocabularySet
	}
	return uc.dictionary.UpdateVocabularySet(ctx, userID, id, title)
}

func (uc *UseCase) DeleteSet(ctx context.Context, userID string, id int64) error {
	if userID == "" || id <= 0 {
		return entity.ErrInvalidVocabularySet
	}
	return uc.dictionary.DeleteVocabularySet(ctx, userID, id)
}

func (uc *UseCase) CreateUserVocabulary(ctx context.Context, userID string, input entity.UserVocabularyInput) (entity.UserVocabulary, error) {
	if userID == "" || input.DictionaryID <= 0 || invalidOptionalID(input.VocabSetID) || invalidOptionalID(input.CaptionID) {
		return entity.UserVocabulary{}, entity.ErrInvalidVocabulary
	}
	return uc.dictionary.CreateUserVocabulary(ctx, userID, input)
}

func (uc *UseCase) ListUserVocabularies(ctx context.Context, userID string, vocabSetID *int64) ([]entity.UserVocabulary, error) {
	if userID == "" || invalidOptionalID(vocabSetID) {
		return nil, entity.ErrInvalidVocabulary
	}
	return uc.dictionary.ListUserVocabularies(ctx, userID, vocabSetID)
}

func (uc *UseCase) UpdateUserVocabulary(ctx context.Context, userID string, id int64, input entity.UserVocabularyUpdate) (entity.UserVocabulary, error) {
	if userID == "" || id <= 0 || invalidOptionalID(input.VocabSetID) || invalidOptionalID(input.CaptionID) {
		return entity.UserVocabulary{}, entity.ErrInvalidVocabulary
	}
	return uc.dictionary.UpdateUserVocabulary(ctx, userID, id, input)
}

func (uc *UseCase) DeleteUserVocabulary(ctx context.Context, userID string, id int64) error {
	if userID == "" || id <= 0 {
		return entity.ErrInvalidVocabulary
	}
	return uc.dictionary.DeleteUserVocabulary(ctx, userID, id)
}

func invalidOptionalID(id *int64) bool { return id != nil && *id <= 0 }
