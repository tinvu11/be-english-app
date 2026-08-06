package channel

import (
	"context"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type repoStub struct {
	created entity.Channel
	deleted int
}

func (*repoStub) ListChannels(context.Context) ([]entity.Channel, error) { return nil, nil }
func (r *repoStub) CreateChannel(_ context.Context, v *entity.Channel) error {
	r.created = *v
	v.ID = 1
	return nil
}
func (*repoStub) UpdateChannel(context.Context, *entity.Channel) error { return nil }
func (r *repoStub) DeleteChannel(_ context.Context, id int) error      { r.deleted = id; return nil }

func TestCreateChannel(t *testing.T) {
	t.Parallel()
	repository := &repoStub{}
	uc := &UseCase{repo: repository}
	result, err := uc.CreateChannel(t.Context(), entity.Channel{ChannelYouTubeID: " UC123 ", ChannelName: " Channel ", IsActive: true})
	require.NoError(t, err)
	assert.Equal(t, "UC123", result.ChannelYouTubeID)
	assert.Equal(t, "Channel", result.ChannelName)
	assert.True(t, result.IsActive)
}

func TestCreateChannelRejectsMissingID(t *testing.T) {
	t.Parallel()
	uc := &UseCase{repo: &repoStub{}}
	_, err := uc.CreateChannel(t.Context(), entity.Channel{ChannelName: "Channel"})
	require.ErrorIs(t, err, entity.ErrInvalidChannel)
}
