// Package level implements proficiency-level business rules.
package level

import (
	"context"
	"strings"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/internal/usecase"
)

type UseCase struct{ repo repo.LevelRepo }

func New(repository repo.LevelRepo) usecase.Level {
	return newTraced(&UseCase{repo: repository})
}

func (uc *UseCase) ListLevels(ctx context.Context, languageID *int) ([]entity.Level, error) {
	if languageID != nil && *languageID <= 0 {
		return nil, entity.ErrInvalidLevel
	}

	return uc.repo.ListLevels(ctx, languageID)
}

func (uc *UseCase) CreateLevel(ctx context.Context, level entity.Level) (entity.Level, error) {
	if err := normalizeAndValidate(&level); err != nil {
		return entity.Level{}, err
	}
	if err := uc.repo.CreateLevel(ctx, &level); err != nil {
		return entity.Level{}, err
	}

	return level, nil
}

func (uc *UseCase) UpdateLevel(ctx context.Context, id int, level entity.Level) (entity.Level, error) {
	level.ID = id
	if err := normalizeAndValidate(&level); err != nil {
		return entity.Level{}, err
	}
	if err := uc.repo.UpdateLevel(ctx, &level); err != nil {
		return entity.Level{}, err
	}

	return level, nil
}

func (uc *UseCase) DeleteLevel(ctx context.Context, id int) error {
	if id <= 0 {
		return entity.ErrInvalidLevel
	}

	return uc.repo.DeleteLevel(ctx, id)
}

func normalizeAndValidate(level *entity.Level) error {
	level.Code = strings.ToUpper(strings.TrimSpace(level.Code))
	if level.ID < 0 || level.Code == "" || len(level.Code) > 20 || level.LanguageID <= 0 {
		return entity.ErrInvalidLevel
	}

	seenLanguages := make(map[int]struct{}, len(level.Translations))
	for index := range level.Translations {
		translation := &level.Translations[index]
		translation.Name = strings.TrimSpace(translation.Name)
		if translation.LanguageID <= 0 || translation.Name == "" || len(translation.Name) > 100 {
			return entity.ErrInvalidLevel
		}
		if _, exists := seenLanguages[translation.LanguageID]; exists {
			return entity.ErrInvalidLevel
		}
		seenLanguages[translation.LanguageID] = struct{}{}
	}

	return nil
}
