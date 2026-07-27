package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	contentuc "github.com/evrone/go-clean-template/internal/usecase/content"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var errDependency = errors.New("database unavailable")

func TestContentListTopicsPagination(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repository := NewMockContentRepo(ctrl)
	repository.EXPECT().
		ListTopics(gomock.Any(), repo.ContentFilter{Limit: 100, Offset: 0}).
		Return([]entity.Topic{{ID: 1}}, 1, nil)

	items, total, err := contentuc.New(repository).ListTopics(context.Background(), 1000, -1)

	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, items, 1)
}

func TestContentCreateTopicConflict(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repository := NewMockContentRepo(ctrl)
	repository.EXPECT().CreateTopic(gomock.Any(), gomock.Any()).Return(entity.ErrContentConflict)

	_, err := contentuc.New(repository).CreateTopic(context.Background(), entity.Topic{Name: "Go", Slug: "go"})

	require.ErrorIs(t, err, entity.ErrContentConflict)
}

func TestContentGetVideoDependencyError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repository := NewMockContentRepo(ctrl)
	repository.EXPECT().GetVideo(gomock.Any(), int64(7)).Return(entity.Video{}, errDependency)

	_, err := contentuc.New(repository).GetVideo(context.Background(), 7)

	require.ErrorIs(t, err, errDependency)
}
