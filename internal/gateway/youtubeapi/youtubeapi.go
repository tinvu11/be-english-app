// Package youtubeapi retrieves public video metadata through YouTube Data API v3.
package youtubeapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/gateway"
)

const videosURL = "https://www.googleapis.com/youtube/v3/videos"

var durationPattern = regexp.MustCompile(`^P(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?)?$`)
var errorReasonPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,63}$`)

type Provider struct {
	audio   gateway.YouTubeProvider
	apiKey  string
	client  *http.Client
	baseURL string
}

// New uses the official API for video previews and retains the existing audio provider.
func New(audio gateway.YouTubeProvider, apiKey string, client *http.Client) gateway.YouTubeProvider {
	return &Provider{audio: audio, apiKey: apiKey, client: client, baseURL: videosURL}
}

func (p *Provider) PreviewVideo(ctx context.Context, youtubeID string) (entity.YouTubeVideoPreview, error) {
	if !validYouTubeID(youtubeID) {
		return entity.YouTubeVideoPreview{}, entity.ErrInvalidVideo
	}
	requestURL, err := url.Parse(p.baseURL)
	if err != nil {
		return entity.YouTubeVideoPreview{}, fmt.Errorf("%w: invalid API URL", entity.ErrYouTubeProviderFailed)
	}
	query := requestURL.Query()
	query.Set("part", "snippet,contentDetails")
	query.Set("id", youtubeID)
	query.Set("key", p.apiKey)
	requestURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), http.NoBody)
	if err != nil {
		return entity.YouTubeVideoPreview{}, fmt.Errorf("%w: create API request: %w", entity.ErrYouTubeProviderFailed, err)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err // The outer error contains the request URL and API key.
		}
		return entity.YouTubeVideoPreview{}, fmt.Errorf("%w: API request: %w", entity.ErrYouTubeProviderFailed, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var apiError struct {
			Error struct {
				Status string `json:"status"`
				Errors []struct {
					Reason string `json:"reason"`
				} `json:"errors"`
			} `json:"error"`
		}
		_ = json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&apiError)
		reason := apiError.Error.Status
		if len(apiError.Error.Errors) > 0 && apiError.Error.Errors[0].Reason != "" {
			reason = apiError.Error.Errors[0].Reason
		}
		if !errorReasonPattern.MatchString(reason) {
			reason = ""
		}
		return entity.YouTubeVideoPreview{}, &entity.YouTubeProviderStatusError{StatusCode: resp.StatusCode, Reason: reason}
	}

	var result struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				Title        string `json:"title"`
				ChannelID    string `json:"channelId"`
				ChannelTitle string `json:"channelTitle"`
				Thumbnails   map[string]struct {
					URL string `json:"url"`
				} `json:"thumbnails"`
			} `json:"snippet"`
			ContentDetails struct {
				Duration string `json:"duration"`
			} `json:"contentDetails"`
		} `json:"items"`
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil {
		return entity.YouTubeVideoPreview{}, fmt.Errorf("%w: decode API response: %w", entity.ErrYouTubeProviderFailed, err)
	}
	if len(result.Items) == 0 {
		return entity.YouTubeVideoPreview{}, entity.ErrVideoNotFound
	}
	item := result.Items[0]
	duration, err := parseDuration(item.ContentDetails.Duration)
	if err != nil {
		return entity.YouTubeVideoPreview{}, fmt.Errorf("%w: invalid video duration: %w", entity.ErrYouTubeProviderFailed, err)
	}
	thumbnail := ""
	for _, size := range []string{"maxres", "standard", "high", "medium", "default"} {
		if value, ok := item.Snippet.Thumbnails[size]; ok && value.URL != "" {
			thumbnail = value.URL
			break
		}
	}
	return entity.YouTubeVideoPreview{
		YouTubeID: item.ID, Title: item.Snippet.Title, ThumbnailURL: thumbnail,
		DurationSeconds: duration, ChannelYouTubeID: item.Snippet.ChannelID,
		ChannelName: item.Snippet.ChannelTitle,
	}, nil
}

func (p *Provider) DownloadAudio(ctx context.Context, youtubeID string) ([]byte, error) {
	return p.audio.DownloadAudio(ctx, youtubeID)
}

func parseDuration(value string) (int, error) {
	parts := durationPattern.FindStringSubmatch(value)
	if parts == nil {
		return 0, fmt.Errorf("unsupported ISO 8601 duration %q", value)
	}
	total := 0
	for index, multiplier := range []int{86400, 3600, 60, 1} {
		if parts[index+1] == "" {
			continue
		}
		amount, err := strconv.Atoi(parts[index+1])
		if err != nil {
			return 0, err
		}
		total += amount * multiplier
	}
	return total, nil
}

func validYouTubeID(value string) bool {
	if len(value) < 6 || len(value) > 50 {
		return false
	}
	for _, char := range value {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '-' || char == '_') {
			return false
		}
	}
	return true
}
