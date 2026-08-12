package deepseek

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslateCaptions(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "Bearer secret", request.Header.Get("Authorization"))
		writer.Header().Set("Content-Type", "application/json")
		_, err := writer.Write([]byte(`{"choices":[{"message":{"content":"{\"items\":[{\"captionId\":7,\"order\":1,\"text\":\"Xin chào\"}]}"},"finish_reason":"stop"}]}`))
		require.NoError(t, err)
	}))
	defer server.Close()
	client := New(Config{BaseURL: server.URL, APIKey: "secret", Model: "test-model"}, server.Client())
	items, err := client.TranslateCaptions(context.Background(), entity.CaptionTranslationRequest{
		SourceLanguage: "en", TargetLanguage: "vi", Items: []entity.CaptionText{{CaptionID: 7, Order: 1, Text: "Hello"}},
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Xin chào", items[0].Text)
}

func TestTranslateCaptionsRejectsInvalidResponse(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, err := writer.Write([]byte(`{"choices":[{"message":{"content":""},"finish_reason":"stop"}]}`))
		require.NoError(t, err)
	}))
	defer server.Close()
	client := New(Config{BaseURL: server.URL, APIKey: "secret", Model: "test-model"}, server.Client())
	_, err := client.TranslateCaptions(context.Background(), entity.CaptionTranslationRequest{})
	assert.ErrorIs(t, err, entity.ErrInvalidTranslation)
}

func TestTranslateCaptionsDisabledWithoutAPIKey(t *testing.T) {
	t.Parallel()
	client := New(Config{}, http.DefaultClient)
	_, err := client.TranslateCaptions(context.Background(), entity.CaptionTranslationRequest{})
	assert.ErrorIs(t, err, entity.ErrTranslationUnavailable)
}

func TestTranslateVocabulary(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "Bearer secret", request.Header.Get("Authorization"))
		writer.Header().Set("Content-Type", "application/json")
		_, err := writer.Write([]byte(`{"choices":[{"message":{"content":"{\"phoneticOrPinyin\":\"/rɪˈmembər/\",\"partOfSpeech\":\"verb\",\"meaning\":\"nhớ\",\"example1Sentence\":\"Remember me.\",\"example1Translation\":\"Hãy nhớ tôi.\",\"example2Sentence\":\"\",\"example2Translation\":\"\"}"},"finish_reason":"stop"}]}`))
		require.NoError(t, err)
	}))
	defer server.Close()
	client := New(Config{BaseURL: server.URL, APIKey: "secret", Model: "test-model"}, server.Client())
	result, err := client.TranslateVocabulary(t.Context(), entity.VocabularyTranslationRequest{
		Word: "remember", SourceLanguageCode: "en", TargetLanguageCode: "vi",
	})
	require.NoError(t, err)
	assert.Equal(t, "nhớ", result.Meaning)
	assert.Equal(t, "verb", result.PartOfSpeech)
}
