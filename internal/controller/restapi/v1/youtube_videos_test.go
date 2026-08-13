package v1

import (
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildOriginalVideoCaptionsIncludesStateAndDictationProgress(t *testing.T) {
	captions := entity.VideoCaptions{VideoID: 10, LanguageID: 1, LanguageCode: "en", Total: 2, Items: []entity.VideoCaptionItem{
		{ID: 100, SentenceOrder: 1, Text: "First"},
		{ID: 200, SentenceOrder: 2, Text: "Second"},
	}}
	state := entity.VideoState{Watched: true, Saved: true, LastPositionSeconds: 45}
	completed := entity.DictationProgressList{Items: []entity.DictationProgress{{CaptionID: 200}}}

	got := buildOriginalVideoCaptions(captions, state, completed)

	assert.Equal(t, state, got.VideoState)
	require.Len(t, got.Items, 2)
	assert.False(t, got.Items[0].DictationCompleted)
	assert.True(t, got.Items[1].DictationCompleted)
}
