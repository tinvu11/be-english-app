package groq

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranscribeSendsFLACAndMapsLanguage(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "Bearer secret", request.Header.Get("Authorization"))
		require.NoError(t, request.ParseMultipartForm(1<<20))
		assert.Equal(t, "whisper-large-v3-turbo", request.FormValue("model"))
		assert.Equal(t, "en", request.FormValue("language"))
		file, header, err := request.FormFile("file")
		require.NoError(t, err)
		defer file.Close()
		assert.Equal(t, "audio.flac", header.Filename)
		data, err := io.ReadAll(file)
		require.NoError(t, err)
		assert.Equal(t, []byte("flac-data"), data)
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"language":"english","segments":[{"start":1.25,"end":2.5,"text":" Hello "}]}`))
	}))
	defer server.Close()

	client := New(Config{BaseURL: server.URL, APIKey: "secret", Model: "whisper-large-v3-turbo"}, server.Client())
	result, err := client.Transcribe(t.Context(), []byte("flac-data"), "en")
	require.NoError(t, err)
	assert.Equal(t, "en", result.LanguageCode)
	require.Len(t, result.Segments, 1)
	assert.Equal(t, 1.25, result.Segments[0].StartSeconds)
}

func TestNormalizeLanguage(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "vi", normalizeLanguage("Vietnamese"))
	assert.Equal(t, "ja", normalizeLanguage("ja"))
	assert.Empty(t, normalizeLanguage("unknown language"))
}
