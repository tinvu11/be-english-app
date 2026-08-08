package video

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseYouTubeID(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"dQw4w9WgXcQ": "dQw4w9WgXcQ",
		"https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=1": "dQw4w9WgXcQ",
		"https://youtu.be/dQw4w9WgXcQ?t=1":                "dQw4w9WgXcQ",
		"https://youtube.com/shorts/dQw4w9WgXcQ":          "dQw4w9WgXcQ",
		"https://youtube.com/embed/dQw4w9WgXcQ":           "dQw4w9WgXcQ",
	}
	for input, expected := range tests {
		actual, err := parseYouTubeID(input)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	}
}

func TestParseYouTubeIDRejectsUnsupportedInput(t *testing.T) {
	t.Parallel()
	inputs := []string{"", "https://example.com/watch?v=dQw4w9WgXcQ", "https://youtube.com/watch?v=bad!id", "javascript:alert(1)"}
	for _, input := range inputs {
		_, err := parseYouTubeID(input)
		assert.Error(t, err)
	}
}
