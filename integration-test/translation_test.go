package integration_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"
)

// HTTP POST: /v1/translation/do-translate.
func TestHTTPDoTranslateV1(t *testing.T) {
	token := registerAndLogin(t)

	tests := []struct {
		description string
		body        string
		expected    int
	}{
		{
			description: "DoTranslate Success auto",
			body: `{
				"destination": "en",
				"original": "текст для перевода",
				"source": "auto"
			}`,
			expected: http.StatusOK,
		},
		{
			description: "DoTranslate Success ru",
			body: `{
				"destination": "en",
				"original": "Текст для перевода",
				"source": "ru"
			}`,
			expected: http.StatusOK,
		},
		{
			description: "DoTranslate Fail",
			body: `{
				"destination": "en",
				"original": "текст для перевода"
			}`,
			expected: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			url := basePathV1 + "/translation/do-translate"
			ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)

			defer cancel()

			resp, err := doAuthenticatedRequest(ctx, http.MethodPost, url, bytes.NewBufferString(tt.body), token)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			defer resp.Body.Close()

			if resp.StatusCode != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, resp.StatusCode)
			}
		})
	}
}

// HTTP GET: /v1/translation/history.
func TestHTTPHistoryV1(t *testing.T) {
	token := registerAndLogin(t)

	// First create a translation so history is non-empty.
	translateURL := basePathV1 + "/translation/do-translate"
	translateBody := `{
		"destination": "en",
		"original": "текст для перевода",
		"source": "auto"
	}`

	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()

	resp, err := doAuthenticatedRequest(ctx, http.MethodPost, translateURL, bytes.NewBufferString(translateBody), token)
	if err != nil {
		t.Fatalf("Failed to create translation: %v", err)
	}

	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d for translation, got %d", http.StatusOK, resp.StatusCode)
	}

	// Now fetch history.
	historyURL := basePathV1 + "/translation/history"

	ctx2, cancel2 := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel2()

	resp, err = doAuthenticatedRequest(ctx2, http.MethodGet, historyURL, http.NoBody, token)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	type historyBody struct {
		History []struct {
			Source      string `json:"source"`
			Destination string `json:"destination"`
			Original    string `json:"original"`
			Translation string `json:"translation"`
		} `json:"history"`
	}

	body := parseJSON[historyBody](t, resp)

	if len(body.History) == 0 {
		t.Error("Expected non-empty history")
	}
}
