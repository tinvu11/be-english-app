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
