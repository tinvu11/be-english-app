// Package channel implements YouTube channel business rules.
package channel

import (
	"context"
	"strings"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/internal/usecase"
)

type UseCase struct{ repo repo.ChannelRepo }

func New(repository repo.ChannelRepo) usecase.Channel { return newTraced(&UseCase{repo: repository}) }

func (uc *UseCase) ListChannels(ctx context.Context) ([]entity.Channel, error) {
	return uc.repo.ListChannels(ctx)
}

func (uc *UseCase) CreateChannel(ctx context.Context, channel entity.Channel) (entity.Channel, error) {
	if err := normalizeAndValidate(&channel); err != nil {
		return entity.Channel{}, err
	}
	if err := uc.repo.CreateChannel(ctx, &channel); err != nil {
		return entity.Channel{}, err
	}
	return channel, nil
}

func (uc *UseCase) UpdateChannel(ctx context.Context, id int, channel entity.Channel) (entity.Channel, error) {
	channel.ID = id
	if err := normalizeAndValidate(&channel); err != nil {
		return entity.Channel{}, err
	}
	if err := uc.repo.UpdateChannel(ctx, &channel); err != nil {
		return entity.Channel{}, err
	}
	return channel, nil
}

func (uc *UseCase) DeleteChannel(ctx context.Context, id int) error {
	if id <= 0 {
		return entity.ErrInvalidChannel
	}
	return uc.repo.DeleteChannel(ctx, id)
}

func normalizeAndValidate(channel *entity.Channel) error {
	channel.ChannelYouTubeID = strings.TrimSpace(channel.ChannelYouTubeID)
	channel.ChannelName = strings.TrimSpace(channel.ChannelName)
	channel.AvatarURL = strings.TrimSpace(channel.AvatarURL)
	if channel.ID < 0 || channel.ChannelYouTubeID == "" || len(channel.ChannelYouTubeID) > 100 ||
		channel.ChannelName == "" || len(channel.ChannelName) > 150 {
		return entity.ErrInvalidChannel
	}
	return nil
}
