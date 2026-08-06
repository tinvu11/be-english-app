// Package topic implements topic business rules.
package topic

import (
	"context"
	"regexp"
	"strings"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/internal/usecase"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type UseCase struct{ repo repo.TopicRepo }

func New(repository repo.TopicRepo) usecase.Topic { return newTraced(&UseCase{repo: repository}) }

func (uc *UseCase) ListTopics(ctx context.Context) ([]entity.Topic, error) {
	return uc.repo.ListTopics(ctx)
}

func (uc *UseCase) CreateTopic(ctx context.Context, topic entity.Topic) (entity.Topic, error) {
	if err := normalizeAndValidate(&topic); err != nil {
		return entity.Topic{}, err
	}
	if err := uc.repo.CreateTopic(ctx, &topic); err != nil {
		return entity.Topic{}, err
	}
	return topic, nil
}

func (uc *UseCase) UpdateTopic(ctx context.Context, id int, topic entity.Topic) (entity.Topic, error) {
	topic.ID = id
	if err := normalizeAndValidate(&topic); err != nil {
		return entity.Topic{}, err
	}
	if err := uc.repo.UpdateTopic(ctx, &topic); err != nil {
		return entity.Topic{}, err
	}
	return topic, nil
}

func (uc *UseCase) SetTopicActive(ctx context.Context, id int, isActive bool) (entity.Topic, error) {
	if id <= 0 {
		return entity.Topic{}, entity.ErrInvalidTopic
	}
	return uc.repo.SetTopicActive(ctx, id, isActive)
}

func (uc *UseCase) DeleteTopic(ctx context.Context, id int) error {
	if id <= 0 {
		return entity.ErrInvalidTopic
	}
	return uc.repo.DeleteTopic(ctx, id)
}

func normalizeAndValidate(topic *entity.Topic) error {
	topic.Slug = strings.ToLower(strings.TrimSpace(topic.Slug))
	topic.IconURL = strings.TrimSpace(topic.IconURL)
	if topic.ID < 0 || !slugPattern.MatchString(topic.Slug) || len(topic.Slug) > 100 || len(topic.Translations) == 0 {
		return entity.ErrInvalidTopic
	}
	seen := make(map[int]struct{}, len(topic.Translations))
	for index := range topic.Translations {
		translation := &topic.Translations[index]
		translation.Name = strings.TrimSpace(translation.Name)
		if translation.LanguageID <= 0 || translation.Name == "" || len(translation.Name) > 100 {
			return entity.ErrInvalidTopic
		}
		if _, exists := seen[translation.LanguageID]; exists {
			return entity.ErrInvalidTopic
		}
		seen[translation.LanguageID] = struct{}{}
	}
	return nil
}
