package level

import (
	"context"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type repoStub struct {
	created entity.Level
	updated entity.Level
	deleted int
	err     error
}

func (stub *repoStub) ListLevels(context.Context, *int) ([]entity.Level, error) {
	return nil, stub.err
}

func (stub *repoStub) CreateLevel(_ context.Context, level *entity.Level) error {
	stub.created = *level
	level.ID = 1

	return stub.err
}

func (stub *repoStub) UpdateLevel(_ context.Context, level *entity.Level) error {
	stub.updated = *level

	return stub.err
}

func (stub *repoStub) DeleteLevel(_ context.Context, id int) error {
	stub.deleted = id

	return stub.err
}

func TestCreateLevelNormalizesData(t *testing.T) {
	t.Parallel()

	repository := &repoStub{}
	uc := &UseCase{repo: repository}
	result, err := uc.CreateLevel(t.Context(), entity.Level{
		Code:       " a1 ",
		LanguageID: 1,
		Translations: []entity.LevelTranslation{
			{LanguageID: 1, Name: " Beginner "},
			{LanguageID: 2, Name: " Sơ cấp "},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "A1", result.Code)
	assert.Equal(t, "Beginner", result.Translations[0].Name)
	assert.Equal(t, result.Code, repository.created.Code)
}

func TestCreateLevelRejectsDuplicateTranslationLanguage(t *testing.T) {
	t.Parallel()

	uc := &UseCase{repo: &repoStub{}}
	_, err := uc.CreateLevel(t.Context(), entity.Level{
		Code:       "A1",
		LanguageID: 1,
		Translations: []entity.LevelTranslation{
			{LanguageID: 1, Name: "Beginner"},
			{LanguageID: 1, Name: "Basic"},
		},
	})

	require.ErrorIs(t, err, entity.ErrInvalidLevel)
}

func TestCreateLevelWithoutTranslations(t *testing.T) {
	t.Parallel()

	repository := &repoStub{}
	uc := &UseCase{repo: repository}
	result, err := uc.CreateLevel(t.Context(), entity.Level{Code: " a2 ", LanguageID: 1})

	require.NoError(t, err)
	assert.Equal(t, "A2", result.Code)
	assert.Empty(t, result.Translations)
	assert.Empty(t, repository.created.Translations)
}

func TestDeleteLevel(t *testing.T) {
	t.Parallel()

	repository := &repoStub{}
	uc := &UseCase{repo: repository}
	require.NoError(t, uc.DeleteLevel(t.Context(), 7))
	assert.Equal(t, 7, repository.deleted)
}
