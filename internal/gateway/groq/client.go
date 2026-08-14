// Package groq implements timestamped speech-to-text through Groq's OpenAI-compatible API.
package groq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/gateway"
)

const maxResponseBytes = 10 << 20

type Config struct {
	BaseURL string
	APIKey  string
	Model   string
}

type Client struct {
	config Config
	http   *http.Client
}

func New(config Config, httpClient *http.Client) gateway.AudioTranscriber {
	config.BaseURL = strings.TrimRight(config.BaseURL, "/")
	return &Client{config: config, http: httpClient}
}

type transcriptionResponse struct {
	Language string `json:"language"`
	Segments []struct {
		Start float64 `json:"start"`
		End   float64 `json:"end"`
		Text  string  `json:"text"`
	} `json:"segments"`
}

func (c *Client) Transcribe(ctx context.Context, audio []byte, languageHint string) (entity.AudioTranscription, error) {
	if strings.TrimSpace(c.config.APIKey) == "" || len(audio) == 0 {
		return entity.AudioTranscription{}, entity.ErrTranscriptionFailed
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("file", "audio.flac")
	if err != nil {
		return entity.AudioTranscription{}, fmt.Errorf("%w: create multipart file", entity.ErrTranscriptionFailed)
	}
	if _, err = file.Write(audio); err != nil {
		return entity.AudioTranscription{}, fmt.Errorf("%w: write multipart file", entity.ErrTranscriptionFailed)
	}
	fields := map[string]string{"model": c.config.Model, "response_format": "verbose_json", "temperature": "0"}
	if hint := strings.ToLower(strings.TrimSpace(languageHint)); hint != "" {
		fields["language"] = hint
	}
	for key, value := range fields {
		if err = writer.WriteField(key, value); err != nil {
			return entity.AudioTranscription{}, fmt.Errorf("%w: write multipart field", entity.ErrTranscriptionFailed)
		}
	}
	if err = writer.Close(); err != nil {
		return entity.AudioTranscription{}, fmt.Errorf("%w: close multipart body", entity.ErrTranscriptionFailed)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.BaseURL+"/openai/v1/audio/transcriptions", &body)
	if err != nil {
		return entity.AudioTranscription{}, fmt.Errorf("%w: create request", entity.ErrTranscriptionFailed)
	}
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := c.http.Do(req)
	if err != nil {
		return entity.AudioTranscription{}, fmt.Errorf("%w: %v", entity.ErrTranscriptionFailed, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
		return entity.AudioTranscription{}, fmt.Errorf("%w: status %d", entity.ErrTranscriptionFailed, resp.StatusCode)
	}
	var decoded transcriptionResponse
	if err = json.NewDecoder(io.LimitReader(resp.Body, maxResponseBytes)).Decode(&decoded); err != nil {
		return entity.AudioTranscription{}, fmt.Errorf("%w: decode response", entity.ErrInvalidTranscription)
	}
	result := entity.AudioTranscription{LanguageCode: normalizeLanguage(decoded.Language), Segments: make([]entity.AudioTranscriptionSegment, 0, len(decoded.Segments))}
	for _, segment := range decoded.Segments {
		result.Segments = append(result.Segments, entity.AudioTranscriptionSegment{StartSeconds: segment.Start, EndSeconds: segment.End, Text: segment.Text})
	}
	if result.LanguageCode == "" {
		result.LanguageCode = strings.ToLower(strings.TrimSpace(languageHint))
	}
	return result, nil
}

func normalizeLanguage(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	aliases := map[string]string{
		"english": "en", "vietnamese": "vi", "japanese": "ja", "korean": "ko",
		"chinese": "zh", "mandarin": "zh", "french": "fr", "german": "de",
		"spanish": "es", "italian": "it", "portuguese": "pt", "russian": "ru",
		"thai": "th", "indonesian": "id", "hindi": "hi", "arabic": "ar",
	}
	if code, ok := aliases[value]; ok {
		return code
	}
	if len(value) == 2 || len(value) == 3 {
		return value
	}
	return ""
}
