package caption

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
