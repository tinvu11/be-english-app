package ytdlp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/gateway"
)

const maxRemoteResponseBytes = 100 << 20

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string, httpClient *http.Client) gateway.YouTubeSubtitleProvider {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), http: httpClient}
}

func (c *Client) PreviewVideo(ctx context.Context, youtubeID string) (entity.YouTubeVideoPreview, error) {
	if !validYouTubeID(youtubeID) {
		return entity.YouTubeVideoPreview{}, entity.ErrInvalidVideo
	}
	var preview entity.YouTubeVideoPreview
	err := c.getJSON(ctx, "/v1/videos/"+url.PathEscape(youtubeID), &preview)
	return preview, err
}

func (c *Client) ListManualSubtitles(ctx context.Context, youtubeID string) ([]entity.YouTubeSubtitleTrack, error) {
	preview, err := c.PreviewVideo(ctx, youtubeID)
	if err != nil {
		return nil, err
	}
	return preview.ManualSubtitleTracks, nil
}

func (c *Client) DownloadManualSubtitle(ctx context.Context, youtubeID, languageCode string) ([]byte, error) {
	if !validYouTubeID(youtubeID) || !validLanguageCode(languageCode) {
		return nil, entity.ErrInvalidCaption
	}
	requestURL := c.baseURL + "/v1/videos/" + url.PathEscape(youtubeID) + "/subtitles/" + url.PathEscape(languageCode)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("%w: create request: %w", entity.ErrSubtitleDownloadFailed, err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", entity.ErrSubtitleDownloadFailed, err)
	}
	defer resp.Body.Close()
	if err = remoteStatusError(resp); err != nil {
		return nil, err
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxRemoteResponseBytes+1))
	if err != nil || len(data) > maxRemoteResponseBytes {
		return nil, fmt.Errorf("%w: invalid service response", entity.ErrSubtitleDownloadFailed)
	}
	return data, nil
}

func (c *Client) DownloadAudio(ctx context.Context, youtubeID string) ([]byte, error) {
	if !validYouTubeID(youtubeID) {
		return nil, entity.ErrInvalidVideo
	}
	requestURL := c.baseURL + "/v1/videos/" + url.PathEscape(youtubeID) + "/audio"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("%w: create request: %w", entity.ErrAudioDownloadFailed, err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", entity.ErrAudioDownloadFailed, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
		return nil, fmt.Errorf("%w: service status %d", entity.ErrAudioDownloadFailed, resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxRemoteResponseBytes+1))
	if err != nil || len(data) > maxRemoteResponseBytes {
		return nil, fmt.Errorf("%w: invalid service response", entity.ErrAudioDownloadFailed)
	}
	return data, nil
}

func (c *Client) getJSON(ctx context.Context, path string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, http.NoBody)
	if err != nil {
		return fmt.Errorf("%w: create request: %w", entity.ErrSubtitleDownloadFailed, err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", entity.ErrSubtitleDownloadFailed, err)
	}
	defer resp.Body.Close()
	if err = remoteStatusError(resp); err != nil {
		return err
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, maxRemoteResponseBytes)).Decode(target); err != nil {
		return fmt.Errorf("%w: decode service response: %w", entity.ErrSubtitleDownloadFailed, err)
	}
	return nil
}

func remoteStatusError(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	switch resp.StatusCode {
	case http.StatusBadRequest:
		return entity.ErrInvalidVideo
	case http.StatusNotFound:
		return entity.ErrManualSubtitleNotFound
	default:
		return fmt.Errorf("%w: service status %d", entity.ErrSubtitleDownloadFailed, resp.StatusCode)
	}
}
