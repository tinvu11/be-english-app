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
	binaryPath   string
	timeout      time.Duration
	maxBytes     int64
	potServerURL string
	cookiesFile  string
}

// NewLocal returns an executable-backed provider.
func NewLocal(binaryPath string, timeout time.Duration, maxBytes int64, potServerURL, cookiesFile string) gateway.YouTubeProvider {
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	if maxBytes <= 0 {
		maxBytes = 10 << 20
	}
	return &Provider{binaryPath: binaryPath, timeout: timeout, maxBytes: maxBytes,
		potServerURL: strings.TrimRight(potServerURL, "/"), cookiesFile: strings.TrimSpace(cookiesFile)}
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

type exportedCookie struct {
	Domain         string  `json:"domain"`
	ExpirationDate float64 `json:"expirationDate"`
	HostOnly       bool    `json:"hostOnly"`
	HTTPOnly       bool    `json:"httpOnly"`
	Name           string  `json:"name"`
	Path           string  `json:"path"`
	Secure         bool    `json:"secure"`
	Session        bool    `json:"session"`
	Value          string  `json:"value"`
}

func (p *Provider) PreviewVideo(ctx context.Context, youtubeID string) (entity.YouTubeVideoPreview, error) {
	if !validYouTubeID(youtubeID) {
		return entity.YouTubeVideoPreview{}, entity.ErrInvalidVideo
	}
	output, err := p.run(ctx, p.previewArgs(youtubeID)...)
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
	_, err = p.run(ctx, p.audioArgs(youtubeID, tempDir)...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", entity.ErrAudioDownloadFailed, err)
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

func (p *Provider) previewArgs(youtubeID string) []string {
	args := []string{"--no-config", "--dump-single-json", "--skip-download", "--no-playlist", "--no-warnings"}
	args = append(args, p.cookieArgs()...)
	args = append(args, p.potArgs()...)
	return append(args, youtubeURL(youtubeID))
}

func (p *Provider) audioArgs(youtubeID, tempDir string) []string {
	args := []string{"--no-config", "--no-playlist", "--no-warnings"}
	args = append(args, p.cookieArgs()...)
	args = append(args, p.potArgs()...)
	return append(args, "--extract-audio", "--audio-format", "flac",
		"--postprocessor-args", "ffmpeg:-ar 16000 -ac 1 -sample_fmt s16",
		"--paths", tempDir, "--output", "audio.%(ext)s", youtubeURL(youtubeID))
}

func (p *Provider) potArgs() []string {
	if p.potServerURL == "" {
		return nil
	}
	return []string{"--js-runtimes", "node", "--remote-components", "ejs:github",
		"--extractor-args", "youtube:player_client=mweb",
		"--extractor-args", "youtubepot-bgutilhttp:base_url=" + p.potServerURL}
}

func (p *Provider) cookieArgs() []string {
	if p.cookiesFile == "" {
		return nil
	}
	return []string{"--cookies", p.cookiesFile}
}

func (p *Provider) run(parent context.Context, args ...string) ([]byte, error) {
	args, cleanup, err := prepareCookies(args)
	if err != nil {
		return nil, fmt.Errorf("%w: prepare YouTube cookies: %w", entity.ErrYouTubeProviderFailed, err)
	}
	defer cleanup()

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
		if diagnostic == "" {
			return nil, fmt.Errorf("%w: %w", entity.ErrYouTubeProviderFailed, err)
		}
		if len(diagnostic) > maxDiagnosticBytes {
			diagnostic = diagnostic[:maxDiagnosticBytes]
		}
		return nil, fmt.Errorf("%w: %s", entity.ErrYouTubeProviderFailed, diagnostic)
	}
	return stdout.Bytes(), nil
}

// yt-dlp writes its cookie jar on exit, so give each invocation a writable copy.
func prepareCookies(args []string) ([]string, func(), error) {
	for i, arg := range args {
		if arg != "--cookies" {
			continue
		}
		if i+1 >= len(args) {
			return nil, nil, fmt.Errorf("missing cookies file path")
		}
		data, err := os.ReadFile(args[i+1])
		if err != nil {
			return nil, nil, err
		}
		data, err = netscapeCookies(data)
		if err != nil {
			return nil, nil, err
		}
		copyFile, err := os.CreateTemp("", "youtube-cookies-*.txt")
		if err != nil {
			return nil, nil, err
		}
		cleanup := func() { _ = os.Remove(copyFile.Name()) }
		if _, err = copyFile.Write(data); err != nil {
			_ = copyFile.Close()
			cleanup()
			return nil, nil, err
		}
		if err = copyFile.Close(); err != nil {
			cleanup()
			return nil, nil, err
		}
		prepared := append([]string(nil), args...)
		prepared[i+1] = copyFile.Name()
		return prepared, cleanup, nil
	}
	return args, func() {}, nil
}

func netscapeCookies(data []byte) ([]byte, error) {
	if !bytes.HasPrefix(bytes.TrimSpace(data), []byte("[")) {
		return data, nil
	}
	var cookies []exportedCookie
	if err := json.Unmarshal(data, &cookies); err != nil {
		return nil, fmt.Errorf("invalid JSON cookie export: %w", err)
	}
	if len(cookies) == 0 {
		return nil, fmt.Errorf("JSON cookie export is empty")
	}
	var output strings.Builder
	output.WriteString("# Netscape HTTP Cookie File\n")
	for _, cookie := range cookies {
		if cookie.Domain == "" || cookie.Name == "" ||
			strings.ContainsAny(cookie.Domain+cookie.Path+cookie.Name+cookie.Value, "\t\r\n") {
			return nil, fmt.Errorf("invalid JSON cookie entry")
		}
		domain := cookie.Domain
		if cookie.HTTPOnly {
			domain = "#HttpOnly_" + domain
		}
		includeSubdomains := !cookie.HostOnly || strings.HasPrefix(cookie.Domain, ".")
		path := cookie.Path
		if path == "" {
			path = "/"
		}
		expires := int64(cookie.ExpirationDate)
		if cookie.Session {
			expires = 0
		}
		fmt.Fprintf(&output, "%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
			domain, netscapeBool(includeSubdomains), path, netscapeBool(cookie.Secure), expires, cookie.Name, cookie.Value)
	}
	return []byte(output.String()), nil
}

func netscapeBool(value bool) string {
	if value {
		return "TRUE"
	}
	return "FALSE"
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
