package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/gateway"
)

const maxResponseBytes = 10 << 20

type Config struct {
	BaseURL    string
	APIKey     string
	Model      string
	MaxRetries int
}

type Client struct {
	config Config
	http   *http.Client
}

func New(config Config, httpClient *http.Client) gateway.CaptionTranslator {
	config.BaseURL = strings.TrimRight(config.BaseURL, "/")
	return &Client{config: config, http: httpClient}
}

type chatRequest struct {
	Model          string         `json:"model"`
	Messages       []chatMessage  `json:"messages"`
	ResponseFormat responseFormat `json:"response_format"`
	Temperature    float64        `json:"temperature"`
	Stream         bool           `json:"stream"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

type translationResponse struct {
	Items []entity.TranslatedCaption `json:"items"`
}

func (c *Client) TranslateCaptions(ctx context.Context, input entity.CaptionTranslationRequest) ([]entity.TranslatedCaption, error) {
	if strings.TrimSpace(c.config.APIKey) == "" {
		return nil, entity.ErrTranslationUnavailable
	}
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("DeepSeek - encode input: %w", err)
	}
	payload := chatRequest{
		Model: c.config.Model,
		Messages: []chatMessage{
			{Role: "system", Content: `Translate subtitle captions faithfully. Return JSON only in the exact shape {"items":[{"captionId":1,"order":1,"text":"translation"}]}. Preserve every captionId and order, return exactly one item per input item, never merge or split items, preserve names and subtitle markup when appropriate, and do not add explanations.`},
			{Role: "user", Content: string(inputJSON)},
		},
		ResponseFormat: responseFormat{Type: "json_object"},
		Temperature:    0.1,
		Stream:         false,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("DeepSeek - encode request: %w", err)
	}
	for attempt := 0; ; attempt++ {
		items, retry, requestErr := c.do(ctx, body)
		if requestErr == nil || !retry || attempt >= c.config.MaxRetries {
			return items, requestErr
		}
		if err = wait(ctx, time.Duration(1<<attempt)*250*time.Millisecond); err != nil {
			return nil, fmt.Errorf("%w: %w", entity.ErrTranslationFailed, err)
		}
	}
}

func (c *Client) do(ctx context.Context, body []byte) ([]entity.TranslatedCaption, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, false, fmt.Errorf("DeepSeek - create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, fmt.Errorf("%w: %w", entity.ErrTranslationFailed, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if _, err = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10)); err != nil {
			return nil, false, fmt.Errorf("%w: discard error response: %w", entity.ErrTranslationFailed, err)
		}
		switch resp.StatusCode {
		case http.StatusTooManyRequests:
			return nil, true, entity.ErrTranslationRateLimited
		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return nil, true, fmt.Errorf("%w: upstream status %d", entity.ErrTranslationFailed, resp.StatusCode)
		default:
			return nil, false, fmt.Errorf("%w: upstream status %d", entity.ErrTranslationFailed, resp.StatusCode)
		}
	}
	var response chatResponse
	if err = json.NewDecoder(io.LimitReader(resp.Body, maxResponseBytes)).Decode(&response); err != nil {
		return nil, false, fmt.Errorf("%w: decode response: %w", entity.ErrInvalidTranslation, err)
	}
	if len(response.Choices) != 1 {
		return nil, true, fmt.Errorf("%w: expected one choice actual=%d", entity.ErrInvalidTranslation, len(response.Choices))
	}
	if response.Choices[0].FinishReason != "stop" {
		return nil, true, fmt.Errorf("%w: finish_reason=%q", entity.ErrInvalidTranslation, response.Choices[0].FinishReason)
	}
	if strings.TrimSpace(response.Choices[0].Message.Content) == "" {
		return nil, true, fmt.Errorf("%w: empty response content", entity.ErrInvalidTranslation)
	}
	var translated translationResponse
	if err = json.Unmarshal([]byte(response.Choices[0].Message.Content), &translated); err != nil {
		return nil, false, fmt.Errorf("%w: decode translated items: %w", entity.ErrInvalidTranslation, err)
	}
	return translated.Items, false, nil
}

func wait(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-timer.C:
		return nil
	}
}

var _ gateway.CaptionTranslator = (*Client)(nil)
