package topic

import (
	"context"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type repoStub struct {
	created entity.Topic
	deleted int
}

func (*repoStub) ListTopics(context.Context) ([]entity.Topic, error) { return nil, nil }
func (r *repoStub) CreateTopic(_ context.Context, v *entity.Topic) error {
	r.created = *v
	v.ID = 1
	return nil
}
func (*repoStub) UpdateTopic(context.Context, *entity.Topic) error { return nil }
func (*repoStub) SetTopicActive(context.Context, int, bool) (entity.Topic, error) {
	return entity.Topic{}, nil
}
func (r *repoStub) DeleteTopic(_ context.Context, id int) error { r.deleted = id; return nil }

func TestCreateTopic(t *testing.T) {
	t.Parallel()
	repository := &repoStub{}
	uc := &UseCase{repo: repository}
	result, err := uc.CreateTopic(t.Context(), entity.Topic{Slug: " Travel ", IsActive: true, Translations: []entity.TopicTranslation{{LanguageID: 1, Name: " Travel "}}})
	require.NoError(t, err)
	assert.Equal(t, "travel", result.Slug)
	assert.True(t, result.IsActive)
	assert.Equal(t, "Travel", result.Translations[0].Name)
}

func TestCreateTopicRejectsDuplicateTranslation(t *testing.T) {
	t.Parallel()
	uc := &UseCase{repo: &repoStub{}}
	_, err := uc.CreateTopic(t.Context(), entity.Topic{Slug: "travel", Translations: []entity.TopicTranslation{{LanguageID: 1, Name: "Travel"}, {LanguageID: 1, Name: "Trips"}}})
	require.ErrorIs(t, err, entity.ErrInvalidTopic)
}
