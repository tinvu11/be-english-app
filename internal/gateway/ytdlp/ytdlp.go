// Package ytdlp implements YouTube subtitle access through the yt-dlp executable.
package ytdlp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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

func New(binaryPath string, timeout time.Duration, maxBytes int64) gateway.YouTubeSubtitleProvider {
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	if maxBytes <= 0 {
		maxBytes = 10 << 20
	}
	return &Provider{binaryPath: binaryPath, timeout: timeout, maxBytes: maxBytes}
}

type metadata struct {
	ID        string                      `json:"id"`
	Title     string                      `json:"title"`
	Thumbnail string                      `json:"thumbnail"`
	Duration  float64                     `json:"duration"`
	ChannelID string                      `json:"channel_id"`
	Channel   string                      `json:"channel"`
	Uploader  string                      `json:"uploader"`
	Subtitles map[string][]subtitleFormat `json:"subtitles"`
}

type subtitleFormat struct {
	Extension string `json:"ext"`
	Name      string `json:"name"`
}

func (p *Provider) ListManualSubtitles(ctx context.Context, youtubeID string) ([]entity.YouTubeSubtitleTrack, error) {
	preview, err := p.PreviewVideo(ctx, youtubeID)
	if err != nil {
		return nil, err
	}
	return preview.ManualSubtitleTracks, nil
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
		return entity.YouTubeVideoPreview{}, fmt.Errorf("%w: invalid metadata", entity.ErrSubtitleDownloadFailed)
	}
	tracks := make([]entity.YouTubeSubtitleTrack, 0, len(info.Subtitles))
	for languageCode, formats := range info.Subtitles {
		track := entity.YouTubeSubtitleTrack{LanguageCode: languageCode}
		seen := make(map[string]struct{}, len(formats))
		for _, format := range formats {
			if track.Name == "" {
				track.Name = format.Name
			}
			if format.Extension != "" {
				if _, exists := seen[format.Extension]; !exists {
					track.Formats = append(track.Formats, format.Extension)
					seen[format.Extension] = struct{}{}
				}
			}
		}
		sort.Strings(track.Formats)
		tracks = append(tracks, track)
	}
	sort.Slice(tracks, func(i, j int) bool { return tracks[i].LanguageCode < tracks[j].LanguageCode })
	channelName := info.Channel
	if channelName == "" {
		channelName = info.Uploader
	}
	return entity.YouTubeVideoPreview{YouTubeID: info.ID, Title: info.Title, ThumbnailURL: info.Thumbnail,
		DurationSeconds: int(info.Duration), ChannelYouTubeID: info.ChannelID, ChannelName: channelName,
		ManualSubtitleTracks: tracks}, nil
}

func (p *Provider) DownloadManualSubtitle(ctx context.Context, youtubeID, languageCode string) ([]byte, error) {
	if !validYouTubeID(youtubeID) || !validLanguageCode(languageCode) {
		return nil, entity.ErrInvalidCaption
	}
	tracks, err := p.ListManualSubtitles(ctx, youtubeID)
	if err != nil {
		return nil, err
	}
	found := false
	for _, track := range tracks {
		if track.LanguageCode == languageCode {
			found = true
			break
		}
	}
	if !found {
		return nil, entity.ErrManualSubtitleNotFound
	}
	tempDir, err := os.MkdirTemp("", "youtube-subtitle-*")
	if err != nil {
		return nil, fmt.Errorf("%w: create temporary directory", entity.ErrSubtitleDownloadFailed)
	}
	defer os.RemoveAll(tempDir)
	_, err = p.run(ctx, "--no-config", "--skip-download", "--no-playlist", "--no-warnings",
		"--write-subs", "--no-write-auto-subs", "--sub-langs", languageCode, "--sub-format", "vtt",
		"--paths", tempDir, "--output", "subtitle.%(ext)s", youtubeURL(youtubeID))
	if err != nil {
		return nil, err
	}
	matches, err := filepath.Glob(filepath.Join(tempDir, "subtitle*.vtt"))
	if err != nil || len(matches) != 1 {
		return nil, entity.ErrManualSubtitleNotFound
	}
	file, err := os.Open(matches[0])
	if err != nil {
		return nil, fmt.Errorf("%w: open subtitle", entity.ErrSubtitleDownloadFailed)
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, p.maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%w: read subtitle", entity.ErrSubtitleDownloadFailed)
	}
	if int64(len(data)) > p.maxBytes {
		return nil, fmt.Errorf("%w: subtitle exceeds configured size", entity.ErrSubtitleDownloadFailed)
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
			return nil, fmt.Errorf("%w: %w", entity.ErrSubtitleDownloadFailed, ctx.Err())
		}
		diagnostic := strings.TrimSpace(stderr.String())
		if len(diagnostic) > maxDiagnosticBytes {
			diagnostic = diagnostic[:maxDiagnosticBytes]
		}
		return nil, fmt.Errorf("%w: %s", entity.ErrSubtitleDownloadFailed, diagnostic)
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

func validLanguageCode(value string) bool {
	if value == "" || len(value) > 35 {
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
