package language

import (
	"context"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type repoStub struct {
	languages []entity.Language
	created   entity.Language
	updated   entity.Language
	err       error
}

func (stub *repoStub) ListLanguages(context.Context) ([]entity.Language, error) {
	return stub.languages, stub.err
}

func (stub *repoStub) GetLanguage(context.Context, int) (entity.Language, error) {
	if len(stub.languages) == 0 {
		return entity.Language{}, stub.err
	}
	return stub.languages[0], stub.err
}

func (stub *repoStub) GetLanguageByCode(context.Context, string) (entity.Language, error) {
	if len(stub.languages) == 0 {
		return entity.Language{}, stub.err
	}
	return stub.languages[0], stub.err
}

func (stub *repoStub) CreateLanguage(_ context.Context, language *entity.Language) error {
	stub.created = *language
	language.ID = 1

	return stub.err
}

func (stub *repoStub) UpdateLanguage(_ context.Context, language *entity.Language) error {
	stub.updated = *language
	language.Code = "en"

	return stub.err
}

func TestCreateLanguage(t *testing.T) {
	t.Parallel()

	repository := &repoStub{}
	uc := &UseCase{repo: repository}

	result, err := uc.CreateLanguage(t.Context(), " EN ", " English ", " 🇬🇧 ", false)
	require.NoError(t, err)
	assert.Equal(t, "en", result.Code)
	assert.Equal(t, "English", result.Name)
	assert.Equal(t, "🇬🇧", result.FlagEmoji)
	assert.True(t, result.IsActive)
	assert.False(t, result.IsLearnable)
	assert.Equal(t, result.Code, repository.created.Code)
}

func TestCreateLanguageRejectsInvalidCode(t *testing.T) {
	t.Parallel()

	uc := &UseCase{repo: &repoStub{}}
	_, err := uc.CreateLanguage(t.Context(), "english!", "English", "🇬🇧", false)

	require.ErrorIs(t, err, entity.ErrInvalidLanguage)
}

func TestUpdateLanguage(t *testing.T) {
	t.Parallel()

	repository := &repoStub{}
	uc := &UseCase{repo: repository}

	result, err := uc.UpdateLanguage(t.Context(), 4, " Vietnamese ", " 🇻🇳 ", false, true)
	require.NoError(t, err)
	assert.Equal(t, 4, result.ID)
	assert.Equal(t, "Vietnamese", result.Name)
	assert.Equal(t, "🇻🇳", result.FlagEmoji)
	assert.False(t, result.IsActive)
	assert.True(t, result.IsLearnable)
	assert.Equal(t, result.ID, repository.updated.ID)
}
