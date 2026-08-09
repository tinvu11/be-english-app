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

func TestParseWebVTT(t *testing.T) {
	t.Parallel()
	data := []byte("WEBVTT\nKind: captions\nLanguage: en\n\n00:01.000 --> 00:03.500 align:start\n<v Speaker><c>Hello</c> &amp; welcome\n\nc2\n00:04.000 --> 00:05.250\nAgain")
	items, err := parseWebVTT(data)
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, 1, items[0].SentenceOrder)
	assert.Equal(t, int64(1000), items[0].StartTimeMS)
	assert.Equal(t, int64(3500), items[0].EndTimeMS)
	assert.Equal(t, "Hello & welcome", items[0].Content)
	assert.Equal(t, int64(5250), items[1].EndTimeMS)
}

func TestParseWebVTTRejectsInvalidData(t *testing.T) {
	t.Parallel()
	tests := [][]byte{
		[]byte("not WebVTT"),
		[]byte("WEBVTT\n\n00:03.000 --> 00:01.000\nText"),
		[]byte("WEBVTT\n\n00:01.000 --> 00:02.000\n"),
	}
	for _, data := range tests {
		_, err := parseWebVTT(data)
		assert.Error(t, err)
	}
}
