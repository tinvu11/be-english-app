package azure_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/gateway/azure"
	"github.com/stretchr/testify/require"
)

func TestAssess(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "en-US", r.URL.Query().Get("language"))
		require.Equal(t, "secret", r.Header.Get("Ocp-Apim-Subscription-Key"))
		encoded := r.Header.Get("Pronunciation-Assessment")
		parameters, err := base64.StdEncoding.DecodeString(encoded)
		require.NoError(t, err)
		var value map[string]any
		require.NoError(t, json.Unmarshal(parameters, &value))
		require.Equal(t, "Good morning", value["referenceText"])
		_, _ = io.WriteString(w, `{"Duration":23000000,"NBest":[{"PronunciationAssessment":{"AccuracyScore":79,"FluencyScore":86,"CompletenessScore":100,"PronScore":82,"ProsodyScore":75},"Words":[{"Word":"Good","PronunciationAssessment":{"AccuracyScore":79,"ErrorType":"None"},"Phonemes":[{"Phoneme":"g","PronunciationAssessment":{"AccuracyScore":90}}]}]}]}`)
	}))
	defer server.Close()

	client := azure.New(azure.Config{Endpoint: server.URL, APIKey: "secret"}, server.Client())
	result, err := client.Assess(context.Background(), entity.PronunciationAssessmentInput{
		Audio: []byte("wav"), ContentType: "audio/wav; codecs=audio/pcm; samplerate=16000",
		ReferenceText: "Good morning", Locale: "en-US",
	})
	require.NoError(t, err)
	require.Equal(t, float64(82), result.PronunciationScore)
	require.Equal(t, int64(2300), result.DurationMS)
	require.Len(t, result.Words, 1)
}
