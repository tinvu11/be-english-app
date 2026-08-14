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

func (uc *UseCase) GetOverview(ctx context.Context, userID string) (entity.VocabularyOverview, error) {
	if strings.TrimSpace(userID) == "" {
		return entity.VocabularyOverview{}, entity.ErrInvalidVocabulary
	}
	languages, err := uc.currentLanguages(ctx, userID)
	if err != nil {
		return entity.VocabularyOverview{}, err
	}
	return uc.dictionary.GetVocabularyOverview(ctx, userID, languages)
}

func (uc *UseCase) CreateSet(ctx context.Context, userID, title, colorHex string) (entity.VocabularySet, error) {
	title = strings.TrimSpace(title)
	colorHex = strings.ToUpper(strings.TrimSpace(colorHex))
	if userID == "" || title == "" || utf8.RuneCountInString(title) > 255 || !validColorHex(colorHex) {
		return entity.VocabularySet{}, entity.ErrInvalidVocabularySet
	}
	languages, err := uc.currentLanguages(ctx, userID)
	if err != nil {
		return entity.VocabularySet{}, err
	}
	return uc.dictionary.CreateVocabularySet(ctx, userID, title, colorHex, languages)
}

func (uc *UseCase) ListSets(ctx context.Context, userID string) ([]entity.VocabularySet, error) {
	if userID == "" {
		return nil, entity.ErrInvalidVocabularySet
	}
	languages, err := uc.currentLanguages(ctx, userID)
	if err != nil {
		return nil, err
	}
	return uc.dictionary.ListVocabularySets(ctx, userID, languages)
}

func (uc *UseCase) UpdateSet(ctx context.Context, userID string, id int64, title, colorHex string) (entity.VocabularySet, error) {
	title = strings.TrimSpace(title)
	colorHex = strings.ToUpper(strings.TrimSpace(colorHex))
	if userID == "" || id <= 0 || title == "" || utf8.RuneCountInString(title) > 255 || !validColorHex(colorHex) {
		return entity.VocabularySet{}, entity.ErrInvalidVocabularySet
	}
	languages, err := uc.currentLanguages(ctx, userID)
	if err != nil {
		return entity.VocabularySet{}, err
	}
	return uc.dictionary.UpdateVocabularySet(ctx, userID, id, title, colorHex, languages)
}

func (uc *UseCase) DeleteSet(ctx context.Context, userID string, id int64) error {
	if userID == "" || id <= 0 {
		return entity.ErrInvalidVocabularySet
	}
	languages, err := uc.currentLanguages(ctx, userID)
	if err != nil {
		return err
	}
	return uc.dictionary.DeleteVocabularySet(ctx, userID, id, languages)
}

func (uc *UseCase) CreateUserVocabulary(ctx context.Context, userID string, input entity.UserVocabularyInput) (entity.UserVocabulary, error) {
	if userID == "" || input.DictionaryID <= 0 || input.VocabSetID <= 0 || invalidOptionalID(input.CaptionID) {
		return entity.UserVocabulary{}, entity.ErrInvalidVocabulary
	}
	languages, err := uc.currentLanguages(ctx, userID)
	if err != nil {
		return entity.UserVocabulary{}, err
	}
	return uc.dictionary.CreateUserVocabulary(ctx, userID, input, languages)
}

func (uc *UseCase) ListUserVocabularies(ctx context.Context, userID string, filter entity.UserVocabularyFilter) ([]entity.UserVocabulary, error) {
	filter.Search = strings.TrimSpace(filter.Search)
	if userID == "" || invalidOptionalID(filter.VocabSetID) || utf8.RuneCountInString(filter.Search) > 150 {
		return nil, entity.ErrInvalidVocabulary
	}
	languages, err := uc.currentLanguages(ctx, userID)
	if err != nil {
		return nil, err
	}
	return uc.dictionary.ListUserVocabularies(ctx, userID, filter, languages)
}

func (uc *UseCase) ListUnlearnedVocabularies(ctx context.Context, userID string, filter entity.UnlearnedVocabularyFilter) ([]entity.UserVocabulary, error) {
	if userID == "" || invalidOptionalID(filter.VocabSetID) || filter.Limit != nil && (*filter.Limit <= 0 || *filter.Limit > 1000) {
		return nil, entity.ErrInvalidVocabulary
	}
	languages, err := uc.currentLanguages(ctx, userID)
	if err != nil {
		return nil, err
	}
	return uc.dictionary.ListUnlearnedVocabularies(ctx, userID, filter, languages)
}

func (uc *UseCase) UpdateUserVocabulary(ctx context.Context, userID string, id int64, input entity.UserVocabularyUpdate) (entity.UserVocabulary, error) {
	if userID == "" || id <= 0 || input.VocabSetID <= 0 || invalidOptionalID(input.CaptionID) {
		return entity.UserVocabulary{}, entity.ErrInvalidVocabulary
	}
	languages, err := uc.currentLanguages(ctx, userID)
	if err != nil {
		return entity.UserVocabulary{}, err
	}
	return uc.dictionary.UpdateUserVocabulary(ctx, userID, id, input, languages)
}

func (uc *UseCase) DeleteUserVocabulary(ctx context.Context, userID string, id int64) error {
	if userID == "" || id <= 0 {
		return entity.ErrInvalidVocabulary
	}
	languages, err := uc.currentLanguages(ctx, userID)
	if err != nil {
		return err
	}
	return uc.dictionary.DeleteUserVocabulary(ctx, userID, id, languages)
}

func (uc *UseCase) currentLanguages(ctx context.Context, userID string) (entity.VocabularyLanguages, error) {
	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return entity.VocabularyLanguages{}, err
	}
	if user.TargetLanguageID == nil || *user.TargetLanguageID <= 0 {
		return entity.VocabularyLanguages{}, entity.ErrTargetLanguageRequired
	}
	if user.NativeLanguageID == nil || *user.NativeLanguageID <= 0 {
		return entity.VocabularyLanguages{}, entity.ErrNativeLanguageRequired
	}
	if *user.TargetLanguageID == *user.NativeLanguageID {
		return entity.VocabularyLanguages{}, entity.ErrInvalidVocabularySet
	}
	return entity.VocabularyLanguages{SourceLanguageID: *user.TargetLanguageID, TargetLanguageID: *user.NativeLanguageID}, nil
}

func invalidOptionalID(id *int64) bool { return id != nil && *id <= 0 }

func validColorHex(color string) bool {
	if len(color) != 7 || color[0] != '#' {
		return false
	}
	for i := 1; i < len(color); i++ {
		if color[i] < '0' || color[i] > '9' {
			if color[i] < 'A' || color[i] > 'F' {
				return false
			}
		}
	}
	return true
}
