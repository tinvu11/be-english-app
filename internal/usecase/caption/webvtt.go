package caption

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/evrone/go-clean-template/internal/entity"
)

var webVTTTag = regexp.MustCompile(`<[^>]*>`)

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
		content := strings.TrimSpace(html.UnescapeString(webVTTTag.ReplaceAllString(strings.Join(contentLines, "\n"), "")))
		if content == "" {
			return nil, fmt.Errorf("empty cue content")
		}
		items = append(items, entity.CaptionInput{SentenceOrder: len(items) + 1, StartTimeMS: start, EndTimeMS: end, Content: content})
	}
	return items, nil
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
