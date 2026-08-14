package caption

import (
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateTranslationBatch(t *testing.T) {
	t.Parallel()
	source := []entity.CaptionText{{CaptionID: 1, Order: 1}, {CaptionID: 2, Order: 2}}
	items, err := validateTranslationBatch(source, []entity.TranslatedCaption{
		{CaptionID: 2, Order: 2, Text: " Hai "},
		{CaptionID: 1, Order: 1, Text: "Một"},
	})
	require.NoError(t, err)
	assert.Equal(t, []entity.CaptionTranslationUpsert{{CaptionID: 2, Text: "Hai"}, {CaptionID: 1, Text: "Một"}}, items)
}

func TestValidateTranslationBatchRejectsProviderDrift(t *testing.T) {
	t.Parallel()
	source := []entity.CaptionText{{CaptionID: 1, Order: 1}}
	tests := [][]entity.TranslatedCaption{
		{},
		{{CaptionID: 99, Order: 1, Text: "text"}},
		{{CaptionID: 1, Order: 2, Text: "text"}},
		{{CaptionID: 1, Order: 1, Text: " "}},
	}
	for _, translated := range tests {
		_, err := validateTranslationBatch(source, translated)
		assert.ErrorIs(t, err, entity.ErrInvalidTranslation)
	}
}

func TestParseSRT(t *testing.T) {
	t.Parallel()
	data := []byte("\ufeff1\r\n00:00:01,000 --> 00:00:03,500\r\nHello\r\nworld\r\n\r\n2\r\n00:00:04.000 --> 00:00:05.250\r\nAgain")
	items, err := parseSRT(data)
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, 1, items[0].SentenceOrder)
	assert.Equal(t, int64(1000), items[0].StartTimeMS)
	assert.Equal(t, int64(3500), items[0].EndTimeMS)
	assert.Equal(t, "Hello\nworld", items[0].Content)
	assert.Equal(t, int64(5250), items[1].EndTimeMS)
}

func TestParseSRTRejectsInvalidData(t *testing.T) {
	t.Parallel()
	tests := [][]byte{
		[]byte("not srt"),
		[]byte("1\n00:00:03,000 --> 00:00:01,000\nText"),
		[]byte("1\n00:61:00,000 --> 00:62:00,000\nText"),
		[]byte("1\n00:00:01,000 --> 00:00:02,000\nText\n\n1\n00:00:03,000 --> 00:00:04,000\nAgain"),
	}
	for _, data := range tests {
		_, err := parseSRT(data)
		assert.Error(t, err)
	}
}

func TestTranscriptionInputsConvertsTimestamps(t *testing.T) {
	t.Parallel()
	items, err := transcriptionInputs(entity.AudioTranscription{LanguageCode: "en", Segments: []entity.AudioTranscriptionSegment{
		{StartSeconds: 1.25, EndSeconds: 3.5, Text: " Hello world "},
	}})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, 1, items[0].SentenceOrder)
	assert.Equal(t, int64(1250), items[0].StartTimeMS)
	assert.Equal(t, int64(3500), items[0].EndTimeMS)
	assert.Equal(t, "Hello world", items[0].Content)
}
