package content

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/internal/usecase"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// UseCase -.
type UseCase struct{ repo repo.ContentRepo }

// New returns a content usecase instrumented with OpenTelemetry tracing spans.
func New(r repo.ContentRepo) usecase.Content { return newTraced(&UseCase{repo: r}) }

func filter(limit, offset int) repo.ContentFilter {
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}

	return repo.ContentFilter{Limit: uint64(limit), Offset: uint64(offset)}
}

// CreateTopic -.
func (uc *UseCase) CreateTopic(ctx context.Context, item entity.Topic) (entity.Topic, error) {
	if err := uc.repo.CreateTopic(ctx, &item); err != nil {
		return entity.Topic{}, fmt.Errorf("ContentUseCase - CreateTopic: %w", err)
	}

	return item, nil
}

// GetTopic -.
func (uc *UseCase) GetTopic(ctx context.Context, id int64) (entity.Topic, error) {
	item, err := uc.repo.GetTopic(ctx, id)

	return item, wrap("GetTopic", err)
}

// ListTopics -.
func (uc *UseCase) ListTopics(ctx context.Context, limit, offset int) ([]entity.Topic, int, error) {
	items, total, err := uc.repo.ListTopics(ctx, filter(limit, offset))

	return items, total, wrap("ListTopics", err)
}

// UpdateTopic -.
func (uc *UseCase) UpdateTopic(ctx context.Context, id int64, item entity.Topic) (entity.Topic, error) {
	item.ID = id
	if err := uc.repo.UpdateTopic(ctx, &item); err != nil {
		return entity.Topic{}, wrap("UpdateTopic", err)
	}

	return uc.GetTopic(ctx, id)
}

// DeleteTopic -.
func (uc *UseCase) DeleteTopic(ctx context.Context, id int64) error {
	return wrap("DeleteTopic", uc.repo.DeleteTopic(ctx, id))
}

// CreateLevel -.
func (uc *UseCase) CreateLevel(ctx context.Context, item entity.Level) (entity.Level, error) {
	if err := uc.repo.CreateLevel(ctx, &item); err != nil {
		return entity.Level{}, wrap("CreateLevel", err)
	}

	return item, nil
}

// GetLevel -.
func (uc *UseCase) GetLevel(ctx context.Context, id int64) (entity.Level, error) {
	item, err := uc.repo.GetLevel(ctx, id)

	return item, wrap("GetLevel", err)
}

// ListLevels -.
func (uc *UseCase) ListLevels(ctx context.Context, limit, offset int) ([]entity.Level, int, error) {
	items, total, err := uc.repo.ListLevels(ctx, filter(limit, offset))

	return items, total, wrap("ListLevels", err)
}

// UpdateLevel -.
func (uc *UseCase) UpdateLevel(ctx context.Context, id int64, item entity.Level) (entity.Level, error) {
	item.ID = id
	if err := uc.repo.UpdateLevel(ctx, &item); err != nil {
		return entity.Level{}, wrap("UpdateLevel", err)
	}

	return uc.GetLevel(ctx, id)
}

// DeleteLevel -.
func (uc *UseCase) DeleteLevel(ctx context.Context, id int64) error {
	return wrap("DeleteLevel", uc.repo.DeleteLevel(ctx, id))
}

// CreateChannel -.
func (uc *UseCase) CreateChannel(ctx context.Context, item entity.Channel) (entity.Channel, error) {
	if err := uc.repo.CreateChannel(ctx, &item); err != nil {
		return entity.Channel{}, wrap("CreateChannel", err)
	}

	return item, nil
}

// GetChannel -.
func (uc *UseCase) GetChannel(ctx context.Context, id int64) (entity.Channel, error) {
	item, err := uc.repo.GetChannel(ctx, id)

	return item, wrap("GetChannel", err)
}

// ListChannels -.
func (uc *UseCase) ListChannels(ctx context.Context, limit, offset int) ([]entity.Channel, int, error) {
	items, total, err := uc.repo.ListChannels(ctx, filter(limit, offset))

	return items, total, wrap("ListChannels", err)
}

// UpdateChannel -.
func (uc *UseCase) UpdateChannel(ctx context.Context, id int64, item entity.Channel) (entity.Channel, error) {
	item.ID = id
	if err := uc.repo.UpdateChannel(ctx, &item); err != nil {
		return entity.Channel{}, wrap("UpdateChannel", err)
	}

	return uc.GetChannel(ctx, id)
}

// DeleteChannel -.
func (uc *UseCase) DeleteChannel(ctx context.Context, id int64) error {
	return wrap("DeleteChannel", uc.repo.DeleteChannel(ctx, id))
}

// CreateVideo -.
func (uc *UseCase) CreateVideo(ctx context.Context, item entity.Video) (entity.Video, error) {
	if err := uc.repo.CreateVideo(ctx, &item); err != nil {
		return entity.Video{}, wrap("CreateVideo", err)
	}

	return item, nil
}

// GetVideo -.
func (uc *UseCase) GetVideo(ctx context.Context, id int64) (entity.Video, error) {
	item, err := uc.repo.GetVideo(ctx, id)

	return item, wrap("GetVideo", err)
}

// ListVideos -.
func (uc *UseCase) ListVideos(ctx context.Context, limit, offset int) ([]entity.Video, int, error) {
	items, total, err := uc.repo.ListVideos(ctx, filter(limit, offset))

	return items, total, wrap("ListVideos", err)
}

// UpdateVideo -.
func (uc *UseCase) UpdateVideo(ctx context.Context, id int64, item entity.Video) (entity.Video, error) {
	item.ID = id
	if err := uc.repo.UpdateVideo(ctx, &item); err != nil {
		return entity.Video{}, wrap("UpdateVideo", err)
	}

	return uc.GetVideo(ctx, id)
}

// DeleteVideo -.
func (uc *UseCase) DeleteVideo(ctx context.Context, id int64) error {
	return wrap("DeleteVideo", uc.repo.DeleteVideo(ctx, id))
}

func wrap(op string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("ContentUseCase - %s: %w", op, err)
}
