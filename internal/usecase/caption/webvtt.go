package caption

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/evrone/go-clean-template/internal/entity"
)

var (
	webVTTTag = regexp.MustCompile(`<[^>]*>`)
	// YouTube captions commonly contain non-speech accessibility cues. Keep this
	// list deliberately narrow so meaningful bracketed text is not discarded.
	nonSpeechCue  = regexp.MustCompile(`(?i)[\[(]\s*(music|applause|laughter|laughing|cheering|cheers|silence|inaudible|noise|background noise|instrumental)\s*[\])]`)
	speakerMarker = regexp.MustCompile(`^\s*>>\s*`)
	captionSpace  = regexp.MustCompile(`[ \t\f\v]+`)
)

func parseWebVTT(data []byte) ([]entity.CaptionInput, error) {
	text := strings.TrimPrefix(string(data), "\ufeff")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || !strings.HasPrefix(strings.TrimSpace(lines[0]), "WEBVTT") {
		return nil, fmt.Errorf("missing WEBVTT header")
	}
	items := make([]entity.CaptionInput, 0)
	index := 1
	for index < len(lines) && strings.TrimSpace(lines[index]) != "" {
		index++
	}
	for index < len(lines) {
		line := strings.TrimSpace(lines[index])
		if line == "" {
			index++
			continue
		}
		if line == "STYLE" || line == "REGION" || strings.HasPrefix(line, "NOTE") {
			for index < len(lines) && strings.TrimSpace(lines[index]) != "" {
				index++
			}
			continue
		}
		if !strings.Contains(line, "-->") {
			index++
			if index >= len(lines) {
				return nil, fmt.Errorf("cue identifier without timestamp")
			}
			line = strings.TrimSpace(lines[index])
		}
		if !strings.Contains(line, "-->") {
			return nil, fmt.Errorf("invalid cue timestamp %q", line)
		}
		times := strings.SplitN(line, "-->", 2)
		start, err := parseVTTTimestamp(strings.TrimSpace(times[0]))
		if err != nil {
			return nil, err
		}
		endFields := strings.Fields(strings.TrimSpace(times[1]))
		if len(endFields) == 0 {
			return nil, fmt.Errorf("missing cue end timestamp")
		}
		end, err := parseVTTTimestamp(endFields[0])
		if err != nil || end < start {
			return nil, fmt.Errorf("invalid cue end timestamp")
		}
		index++
		contentLines := make([]string, 0, 2)
		for index < len(lines) && strings.TrimSpace(lines[index]) != "" {
			contentLines = append(contentLines, strings.TrimSpace(lines[index]))
			index++
		}
		content := cleanYouTubeCaption(strings.Join(contentLines, "\n"))
		if content == "" {
			// A cue containing only an accessibility marker such as [Music] has
			// no translatable speech and should not be persisted.
			continue
		}
		items = append(items, entity.CaptionInput{SentenceOrder: len(items) + 1, StartTimeMS: start, EndTimeMS: end, Content: content})
	}
	return items, nil
}

func cleanYouTubeCaption(raw string) string {
	content := html.UnescapeString(webVTTTag.ReplaceAllString(raw, ""))
	lines := strings.Split(content, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = speakerMarker.ReplaceAllString(line, "")
		line = nonSpeechCue.ReplaceAllString(line, " ")
		line = strings.TrimSpace(captionSpace.ReplaceAllString(line, " "))
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}
	return strings.Join(cleaned, " ")
}

func parseVTTTimestamp(raw string) (int64, error) {
	parts := strings.Split(raw, ":")
	if len(parts) == 2 {
		raw = "00:" + raw
	} else if len(parts) != 3 {
		return 0, fmt.Errorf("invalid WebVTT timestamp %q", raw)
	}
	return parseTimestamp(raw)
}
