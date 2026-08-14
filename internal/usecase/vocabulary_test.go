package usecase_test

import (
	"context"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/usecase/vocabulary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type vocabularyTranslatorStub struct {
	called *bool
	result entity.VocabularyTranslation
}

func (s vocabularyTranslatorStub) TranslateVocabulary(context.Context, entity.VocabularyTranslationRequest) (entity.VocabularyTranslation, error) {
	*s.called = true
	return s.result, nil
}

func TestTranslateWordReusesDictionaryEntry(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	dictionary := NewMockDictionaryRepo(ctrl)
	users := NewMockUserRepo(ctrl)
	languages := NewMockLanguageRepo(ctrl)
	sourceID, targetID := 1, 2
	entry := entity.DictionaryEntry{ID: 9, Word: "remember", SourceLanguageID: sourceID,
		TargetLanguageID: targetID, Meaning: "nhớ"}
	called := false
	users.EXPECT().GetByID(gomock.Any(), "user-1").Return(entity.User{TargetLanguageID: &sourceID, NativeLanguageID: &targetID}, nil)
	dictionary.EXPECT().FindDictionaryEntry(gomock.Any(), "remember", sourceID, targetID).Return(entry, nil)

	result, err := vocabulary.New(dictionary, users, languages, vocabularyTranslatorStub{called: &called}).
		TranslateWord(t.Context(), "user-1", " Remember ")
	require.NoError(t, err)
	assert.True(t, result.Reused)
	assert.Equal(t, entry, result.Entry)
	assert.False(t, called)
}

func TestTranslateWordCallsAIAndStoresMissingEntry(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	dictionary := NewMockDictionaryRepo(ctrl)
	users := NewMockUserRepo(ctrl)
	languages := NewMockLanguageRepo(ctrl)
	sourceID, targetID := 1, 2
	source := entity.Language{ID: sourceID, Code: "en", Name: "English", IsActive: true}
	target := entity.Language{ID: targetID, Code: "vi", Name: "Vietnamese", IsActive: true}
	translation := entity.VocabularyTranslation{PhoneticOrPinyin: "/rɪˈmembər/", PartOfSpeech: "verb",
		Meaning: "nhớ", Example1Sentence: "Remember me.", Example1Translation: "Hãy nhớ tôi."}
	stored := entity.DictionaryEntry{ID: 10, Word: "remember", SourceLanguageID: sourceID,
		TargetLanguageID: targetID, Meaning: "nhớ"}
	called := false
	users.EXPECT().GetByID(gomock.Any(), "user-1").Return(entity.User{TargetLanguageID: &sourceID, NativeLanguageID: &targetID}, nil)
	dictionary.EXPECT().FindDictionaryEntry(gomock.Any(), "remember", sourceID, targetID).Return(entity.DictionaryEntry{}, entity.ErrDictionaryNotFound)
	languages.EXPECT().GetLanguage(gomock.Any(), sourceID).Return(source, nil)
	languages.EXPECT().GetLanguage(gomock.Any(), targetID).Return(target, nil)
	dictionary.EXPECT().UpsertDictionaryEntry(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, input entity.DictionaryInput) (entity.DictionaryEntry, bool, error) {
			assert.Equal(t, "nhớ", input.Meaning)
			return stored, true, nil
		})

	result, err := vocabulary.New(dictionary, users, languages, vocabularyTranslatorStub{called: &called, result: translation}).
		TranslateWord(t.Context(), "user-1", "remember")
	require.NoError(t, err)
	assert.True(t, called)
	assert.False(t, result.Reused)
	assert.Equal(t, stored, result.Entry)
}

func TestVocabularySetAndSavedWordOperations(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	dictionary := NewMockDictionaryRepo(ctrl)
	users := NewMockUserRepo(ctrl)
	languages := NewMockLanguageRepo(ctrl)
	uc := vocabulary.New(dictionary, users, languages, vocabularyTranslatorStub{})
	sourceID, targetID := 1, 2
	pair := entity.VocabularyLanguages{SourceLanguageID: sourceID, TargetLanguageID: targetID}
	set := entity.VocabularySet{ID: 2, Title: "Travel", SourceLanguageID: sourceID, TargetLanguageID: targetID}
	word := entity.UserVocabulary{ID: 3, VocabSetID: set.ID}
	overview := entity.VocabularyOverview{TotalWords: 3, LearnedWords: 1, UnlearnedWords: 2}
	users.EXPECT().GetByID(gomock.Any(), "user-1").Return(entity.User{TargetLanguageID: &sourceID, NativeLanguageID: &targetID}, nil).Times(6)
	dictionary.EXPECT().GetVocabularyOverview(gomock.Any(), "user-1", pair).Return(overview, nil)
	dictionary.EXPECT().CreateVocabularySet(gomock.Any(), "user-1", "Travel", pair).Return(set, nil)
	dictionary.EXPECT().ListVocabularySets(gomock.Any(), "user-1", pair).Return([]entity.VocabularySet{set}, nil)
	dictionary.EXPECT().CreateUserVocabulary(gomock.Any(), "user-1", entity.UserVocabularyInput{DictionaryID: 4, VocabSetID: set.ID}, pair).Return(word, nil)
	dictionary.EXPECT().UpdateUserVocabulary(gomock.Any(), "user-1", int64(3), entity.UserVocabularyUpdate{VocabSetID: set.ID, IsLearned: true}, pair).Return(word, nil)
	dictionary.EXPECT().DeleteUserVocabulary(gomock.Any(), "user-1", int64(3), pair).Return(nil)

	result, err := uc.GetOverview(t.Context(), "user-1")
	require.NoError(t, err)
	assert.Equal(t, overview, result)

	created, err := uc.CreateSet(t.Context(), "user-1", " Travel ")
	require.NoError(t, err)
	assert.Equal(t, set, created)
	sets, err := uc.ListSets(t.Context(), "user-1")
	require.NoError(t, err)
	assert.Len(t, sets, 1)
	saved, err := uc.CreateUserVocabulary(t.Context(), "user-1", entity.UserVocabularyInput{DictionaryID: 4, VocabSetID: set.ID})
	require.NoError(t, err)
	assert.Equal(t, word, saved)
	_, err = uc.UpdateUserVocabulary(t.Context(), "user-1", 3, entity.UserVocabularyUpdate{VocabSetID: set.ID, IsLearned: true})
	require.NoError(t, err)
	require.NoError(t, uc.DeleteUserVocabulary(t.Context(), "user-1", 3))

	_, err = uc.CreateUserVocabulary(t.Context(), "user-1", entity.UserVocabularyInput{DictionaryID: 4})
	assert.ErrorIs(t, err, entity.ErrInvalidVocabulary)
	_, err = uc.UpdateUserVocabulary(t.Context(), "user-1", 3, entity.UserVocabularyUpdate{IsLearned: true})
	assert.ErrorIs(t, err, entity.ErrInvalidVocabulary)
}
