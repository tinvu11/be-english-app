// Package azure implements scripted pronunciation assessment through Azure Speech REST API.
package azure

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/gateway"
)

const maxResponseBytes = 10 << 20

type Config struct {
	Endpoint string
	APIKey   string
}

type Client struct {
	config Config
	http   *http.Client
}

func New(config Config, httpClient *http.Client) gateway.PronunciationAssessor {
	config.Endpoint = strings.TrimRight(config.Endpoint, "/")
	return &Client{config: config, http: httpClient}
}

type response struct {
	Duration int64 `json:"Duration"`
	NBest    []struct {
		AccuracyScore           *float64 `json:"AccuracyScore"`
		FluencyScore            *float64 `json:"FluencyScore"`
		CompletenessScore       *float64 `json:"CompletenessScore"`
		PronScore               *float64 `json:"PronScore"`
		ProsodyScore            *float64 `json:"ProsodyScore"`
		PronunciationAssessment struct {
			AccuracyScore     *float64 `json:"AccuracyScore"`
			FluencyScore      *float64 `json:"FluencyScore"`
			CompletenessScore *float64 `json:"CompletenessScore"`
			PronScore         *float64 `json:"PronScore"`
			ProsodyScore      *float64 `json:"ProsodyScore"`
		} `json:"PronunciationAssessment"`
		Words []struct {
			Word                    string   `json:"Word"`
			AccuracyScore           *float64 `json:"AccuracyScore"`
			ErrorType               string   `json:"ErrorType"`
			PronunciationAssessment struct {
				AccuracyScore *float64 `json:"AccuracyScore"`
				ErrorType     string   `json:"ErrorType"`
			} `json:"PronunciationAssessment"`
			Phonemes json.RawMessage `json:"Phonemes"`
		} `json:"Words"`
	} `json:"NBest"`
}

func (c *Client) Assess(ctx context.Context, input entity.PronunciationAssessmentInput) (entity.PronunciationAssessment, error) {
	if c.config.APIKey == "" || c.config.Endpoint == "" {
		return entity.PronunciationAssessment{}, entity.ErrPronunciationUnavailable
	}
	if len(input.Audio) == 0 || strings.TrimSpace(input.ReferenceText) == "" || strings.TrimSpace(input.Locale) == "" {
		return entity.PronunciationAssessment{}, entity.ErrInvalidShadowingAudio
	}
	parameters, err := json.Marshal(map[string]any{
		"referenceText":           input.ReferenceText,
		"gradingSystem":           "HundredMark",
		"granularity":             "Phoneme",
		"dimension":               "Comprehensive",
		"enableMiscue":            true,
		"enableProsodyAssessment": true,
	})
	if err != nil {
		return entity.PronunciationAssessment{}, fmt.Errorf("%w: encode parameters", entity.ErrPronunciationFailed)
	}
	endpoint := c.config.Endpoint + "/stt/speech/recognition/conversation/cognitiveservices/v1"
	query := url.Values{"language": {input.Locale}, "format": {"detailed"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"?"+query.Encode(), bytes.NewReader(input.Audio))
	if err != nil {
		return entity.PronunciationAssessment{}, fmt.Errorf("%w: create request", entity.ErrPronunciationFailed)
	}
	req.Header.Set("Ocp-Apim-Subscription-Key", c.config.APIKey)
	req.Header.Set("Content-Type", input.ContentType)
	req.Header.Set("Pronunciation-Assessment", base64.StdEncoding.EncodeToString(parameters))
	resp, err := c.http.Do(req)
	if err != nil {
		return entity.PronunciationAssessment{}, fmt.Errorf("%w: %v", entity.ErrPronunciationFailed, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil || len(raw) > maxResponseBytes {
		return entity.PronunciationAssessment{}, fmt.Errorf("%w: read response", entity.ErrPronunciationFailed)
	}
	if resp.StatusCode != http.StatusOK {
		return entity.PronunciationAssessment{}, fmt.Errorf("%w: status %d", entity.ErrPronunciationFailed, resp.StatusCode)
	}
	var decoded response
	if err = json.Unmarshal(raw, &decoded); err != nil || len(decoded.NBest) == 0 {
		return entity.PronunciationAssessment{}, fmt.Errorf("%w: decode response", entity.ErrInvalidPronunciation)
	}
	best := decoded.NBest[0]
	accuracyScore := score(best.AccuracyScore, best.PronunciationAssessment.AccuracyScore)
	fluencyScore := score(best.FluencyScore, best.PronunciationAssessment.FluencyScore)
	completenessScore := score(best.CompletenessScore, best.PronunciationAssessment.CompletenessScore)
	pronunciationScore := score(best.PronScore, best.PronunciationAssessment.PronScore)
	prosodyScore := optionalScore(best.ProsodyScore, best.PronunciationAssessment.ProsodyScore)
	result := entity.PronunciationAssessment{
		Provider: entity.ShadowingProviderAzure, AccuracyScore: accuracyScore,
		FluencyScore: fluencyScore, CompletenessScore: completenessScore,
		PronunciationScore: pronunciationScore, ProsodyScore: prosodyScore,
		DurationMS: decoded.Duration / 10000, ProviderResponse: append(json.RawMessage(nil), raw...),
		Words: make([]entity.ShadowingWordResult, 0, len(best.Words)),
	}
	for index, word := range best.Words {
		accuracyScore := score(word.AccuracyScore, word.PronunciationAssessment.AccuracyScore)
		errorType := word.ErrorType
		if errorType == "" {
			errorType = word.PronunciationAssessment.ErrorType
		}
		result.Words = append(result.Words, entity.ShadowingWordResult{Order: index, Word: word.Word,
			AccuracyScore: accuracyScore, ErrorType: errorType,
			Phonemes: word.Phonemes})
	}
	return result, nil
}

func score(primary, fallback *float64) float64 {
	if primary != nil {
		return *primary
	}
	if fallback != nil {
		return *fallback
	}
	return 0
}

func optionalScore(primary, fallback *float64) *float64 {
	if primary != nil {
		return primary
	}
	return fallback
}
