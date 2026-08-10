// Package language implements language catalog business rules.
package language

import (
	"context"
	"regexp"
	"strings"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/internal/usecase"
)

var codePattern = regexp.MustCompile(`^[a-z]{2,3}(-[a-z0-9]{2,8})*$`)

type UseCase struct {
	repo repo.LanguageRepo
}

func New(repository repo.LanguageRepo) usecase.Language {
	return newTraced(&UseCase{repo: repository})
}

func (uc *UseCase) ListLanguages(ctx context.Context) ([]entity.Language, error) {
	return uc.repo.ListLanguages(ctx)
}

func (uc *UseCase) CreateLanguage(ctx context.Context, code, name, flagEmoji string, isLearnable bool) (entity.Language, error) {
	code = strings.ToLower(strings.TrimSpace(code))
	name = strings.TrimSpace(name)
	flagEmoji = strings.TrimSpace(flagEmoji)
	if !codePattern.MatchString(code) || name == "" || len(name) > 50 || flagEmoji == "" || len([]rune(flagEmoji)) > 16 {
		return entity.Language{}, entity.ErrInvalidLanguage
	}

	language := entity.Language{Code: code, Name: name, FlagEmoji: flagEmoji, IsLearnable: isLearnable, IsActive: true}
	if err := uc.repo.CreateLanguage(ctx, &language); err != nil {
		return entity.Language{}, err
	}

	return language, nil
}

func (uc *UseCase) UpdateLanguage(ctx context.Context, id int, name, flagEmoji string, isActive, isLearnable bool) (entity.Language, error) {
	name = strings.TrimSpace(name)
	flagEmoji = strings.TrimSpace(flagEmoji)
	if id <= 0 || name == "" || len(name) > 50 || flagEmoji == "" || len([]rune(flagEmoji)) > 16 {
		return entity.Language{}, entity.ErrInvalidLanguage
	}

	language := entity.Language{ID: id, Name: name, FlagEmoji: flagEmoji, IsActive: isActive, IsLearnable: isLearnable}
	if err := uc.repo.UpdateLanguage(ctx, &language); err != nil {
		return entity.Language{}, err
	}

	return language, nil
}
