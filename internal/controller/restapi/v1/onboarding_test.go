package v1

import (
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildOnboardingResponse(t *testing.T) {
	languageID := 2
	languages := []entity.Language{
		{ID: 1, Code: "en", Name: "English", FlagEmoji: "🇬🇧", IsLearnable: true, IsActive: true},
		{ID: 2, Code: "vi", Name: "Vietnamese", FlagEmoji: "🇻🇳", IsActive: true},
		{ID: 3, Code: "ja", Name: "Japanese", FlagEmoji: "🇯🇵", IsLearnable: true, IsActive: false},
	}
	topics := []entity.Topic{
		{ID: 1, Slug: "travel", IconURL: "travel.svg", IsActive: true, Translations: []entity.TopicTranslation{{LanguageID: 1, Name: "Travel"}, {LanguageID: 2, Name: "Du lịch"}}},
		{ID: 2, Slug: "inactive", IsActive: false, Translations: []entity.TopicTranslation{{LanguageID: 2, Name: "Ẩn"}}},
		{ID: 3, Slug: "missing-translation", IsActive: true, Translations: []entity.TopicTranslation{{LanguageID: 1, Name: "Missing"}}},
	}

	got := buildOnboardingResponse(languages, topics, &languageID)

	require.Len(t, got.LearningLanguages, 1)
	assert.Equal(t, "🇬🇧", got.LearningLanguages[0].Flag)
	require.Len(t, got.TranslationLanguages, 1)
	assert.Equal(t, "vi", got.TranslationLanguages[0].Code)
	require.Len(t, got.Topics, 1)
	assert.Equal(t, "Du lịch", got.Topics[0].Name)
	assert.Equal(t, "travel.svg", got.Topics[0].Icon)
}

func TestOptionalPositiveInt(t *testing.T) {
	value, err := optionalPositiveInt("2")
	require.NoError(t, err)
	require.NotNil(t, value)
	assert.Equal(t, 2, *value)

	value, err = optionalPositiveInt("")
	require.NoError(t, err)
	assert.Nil(t, value)

	for _, invalid := range []string{"0", "-1", "abc"} {
		_, err = optionalPositiveInt(invalid)
		assert.Error(t, err)
	}
}
