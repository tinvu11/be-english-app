// Package ytdlp implements local and remote YouTube metadata and audio providers.
package ytdlp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/gateway"
	"github.com/goccy/go-json"
)

const maxDiagnosticBytes = 4096

type Provider struct {
	binaryPath string
	timeout    time.Duration
	maxBytes   int64
}

// NewLocal returns an executable-backed provider.
func NewLocal(binaryPath string, timeout time.Duration, maxBytes int64) gateway.YouTubeProvider {
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	if maxBytes <= 0 {
		maxBytes = 10 << 20
	}
	return &Provider{binaryPath: binaryPath, timeout: timeout, maxBytes: maxBytes}
}

type metadata struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Thumbnail string  `json:"thumbnail"`
	Duration  float64 `json:"duration"`
	ChannelID string  `json:"channel_id"`
	Channel   string  `json:"channel"`
	Uploader  string  `json:"uploader"`
}

func (p *Provider) PreviewVideo(ctx context.Context, youtubeID string) (entity.YouTubeVideoPreview, error) {
	if !validYouTubeID(youtubeID) {
		return entity.YouTubeVideoPreview{}, entity.ErrInvalidVideo
	}
	output, err := p.run(ctx, "--no-config", "--dump-single-json", "--skip-download", "--no-playlist",
		"--no-warnings", youtubeURL(youtubeID))
	if err != nil {
		return entity.YouTubeVideoPreview{}, err
	}
	var info metadata
	if err = json.Unmarshal(output, &info); err != nil {
		return entity.YouTubeVideoPreview{}, fmt.Errorf("%w: invalid metadata", entity.ErrYouTubeProviderFailed)
	}
	channelName := info.Channel
	if channelName == "" {
		channelName = info.Uploader
	}
	return entity.YouTubeVideoPreview{YouTubeID: info.ID, Title: info.Title, ThumbnailURL: info.Thumbnail,
		DurationSeconds: int(info.Duration), ChannelYouTubeID: info.ChannelID, ChannelName: channelName}, nil
}

// DownloadAudio extracts the first audio track and normalizes it to 16 kHz mono FLAC.
func (p *Provider) DownloadAudio(ctx context.Context, youtubeID string) ([]byte, error) {
	if !validYouTubeID(youtubeID) {
		return nil, entity.ErrInvalidVideo
	}
	tempDir, err := os.MkdirTemp("", "youtube-audio-*")
	if err != nil {
		return nil, fmt.Errorf("%w: create temporary directory", entity.ErrAudioDownloadFailed)
	}
	defer os.RemoveAll(tempDir)
	_, err = p.run(ctx, "--no-config", "--no-playlist", "--no-warnings", "--extract-audio",
		"--audio-format", "flac", "--postprocessor-args", "ffmpeg:-ar 16000 -ac 1 -sample_fmt s16",
		"--paths", tempDir, "--output", "audio.%(ext)s", youtubeURL(youtubeID))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", entity.ErrAudioDownloadFailed, err)
	}
	path := filepath.Join(tempDir, "audio.flac")
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%w: open normalized audio", entity.ErrAudioDownloadFailed)
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, p.maxBytes+1))
	if err != nil || int64(len(data)) > p.maxBytes {
		return nil, fmt.Errorf("%w: normalized audio exceeds limit", entity.ErrAudioDownloadFailed)
	}
	return data, nil
}

func (p *Provider) run(parent context.Context, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(parent, p.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, p.binaryPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("%w: %w", entity.ErrYouTubeProviderFailed, ctx.Err())
		}
		diagnostic := strings.TrimSpace(stderr.String())
		if len(diagnostic) > maxDiagnosticBytes {
			diagnostic = diagnostic[:maxDiagnosticBytes]
		}
		return nil, fmt.Errorf("%w: %s", entity.ErrYouTubeProviderFailed, diagnostic)
	}
	return stdout.Bytes(), nil
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

func youtubeURL(id string) string { return "https://www.youtube.com/watch?v=" + id }
