package caption

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/internal/usecase"
)

const maxImportCaptions = 10000
const maxCaptionPageSize = 200

type UseCase struct{ repo repo.CaptionRepo }

func New(repository repo.CaptionRepo) usecase.Caption { return newTraced(&UseCase{repo: repository}) }

func (uc *UseCase) ListCaptions(ctx context.Context, videoID int64, filter entity.CaptionFilter) (entity.CaptionList, error) {
	if videoID <= 0 || filter.Limit <= 0 || filter.Limit > maxCaptionPageSize || filter.Offset < 0 {
		return entity.CaptionList{}, entity.ErrInvalidCaption
	}
	return uc.repo.ListCaptions(ctx, videoID, filter)
}

func (uc *UseCase) CreateCaption(ctx context.Context, videoID int64, input entity.CaptionInput) (entity.Caption, error) {
	if videoID <= 0 || normalizeAndValidate(&input) != nil {
		return entity.Caption{}, entity.ErrInvalidCaption
	}
	return uc.repo.CreateCaption(ctx, videoID, input)
}

func (uc *UseCase) ImportSRT(ctx context.Context, videoID int64, original []byte, translationFiles []entity.SRTTranslationFile) ([]entity.Caption, error) {
	if videoID <= 0 || len(original) == 0 {
		return nil, entity.ErrInvalidSRT
	}
	inputs, err := parseSRT(original)
	if err != nil {
		return nil, fmt.Errorf("%w: original file: %v", entity.ErrInvalidSRT, err)
	}
	if len(inputs) == 0 || len(inputs) > maxImportCaptions {
		return nil, fmt.Errorf("%w: original file has %d cues; allowed range is 1-%d", entity.ErrInvalidSRT, len(inputs), maxImportCaptions)
	}
	seenLanguages := make(map[int]struct{}, len(translationFiles))
	for _, file := range translationFiles {
		if file.LanguageID <= 0 || len(file.Data) == 0 {
			return nil, fmt.Errorf("%w: translation language_id=%d is empty or invalid", entity.ErrInvalidSRT, file.LanguageID)
		}
		if _, exists := seenLanguages[file.LanguageID]; exists {
			return nil, entity.ErrInvalidCaption
		}
		seenLanguages[file.LanguageID] = struct{}{}
		translated, parseErr := parseSRT(file.Data)
		if parseErr != nil {
			return nil, fmt.Errorf("%w: translation language_id=%d: %v", entity.ErrInvalidSRT, file.LanguageID, parseErr)
		}
		if len(translated) != len(inputs) {
			return nil, fmt.Errorf("%w: original=%d translation language_id=%d has=%d", entity.ErrCaptionTranslationCount, len(inputs), file.LanguageID, len(translated))
		}
		for index := range inputs {
			inputs[index].Translations = append(inputs[index].Translations, entity.CaptionTranslationInput{
				LanguageID: file.LanguageID, Text: translated[index].Content,
			})
		}
	}
	return uc.repo.ImportCaptions(ctx, videoID, inputs)
}

func (uc *UseCase) UpdateCaption(ctx context.Context, videoID, captionID int64, input entity.CaptionInput) (entity.Caption, error) {
	if videoID <= 0 || captionID <= 0 || normalizeAndValidate(&input) != nil {
		return entity.Caption{}, entity.ErrInvalidCaption
	}
	return uc.repo.UpdateCaption(ctx, videoID, captionID, input)
}

func (uc *UseCase) DeleteCaption(ctx context.Context, videoID, captionID int64) error {
	if videoID <= 0 || captionID <= 0 {
		return entity.ErrInvalidCaption
	}
	return uc.repo.DeleteCaption(ctx, videoID, captionID)
}

func normalizeAndValidate(input *entity.CaptionInput) error {
	input.Content = strings.TrimSpace(input.Content)
	input.PinyinOrFurigana = strings.TrimSpace(input.PinyinOrFurigana)
	if input.SentenceOrder < 0 || input.StartTimeMS < 0 || input.EndTimeMS < input.StartTimeMS || input.Content == "" {
		return entity.ErrInvalidCaption
	}
	seen := make(map[int]struct{}, len(input.Translations))
	for index := range input.Translations {
		input.Translations[index].Text = strings.TrimSpace(input.Translations[index].Text)
		translation := input.Translations[index]
		if translation.LanguageID <= 0 || translation.Text == "" {
			return entity.ErrInvalidCaption
		}
		if _, exists := seen[translation.LanguageID]; exists {
			return entity.ErrInvalidCaption
		}
		seen[translation.LanguageID] = struct{}{}
	}
	return nil
}

func parseSRT(data []byte) ([]entity.CaptionInput, error) {
	text := strings.TrimPrefix(string(data), "\ufeff")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	blocks := strings.Split(strings.TrimSpace(text), "\n\n")
	result := make([]entity.CaptionInput, 0, len(blocks))
	seenOrders := make(map[int]struct{}, len(blocks))
	for blockIndex, block := range blocks {
		cuePosition := blockIndex + 1
		lines := strings.Split(strings.TrimSpace(block), "\n")
		if len(lines) < 3 {
			return nil, fmt.Errorf("cue position %d: expected order, timestamp and content", cuePosition)
		}
		order, err := strconv.Atoi(strings.TrimSpace(lines[0]))
		if err != nil || order < 0 {
			return nil, fmt.Errorf("cue position %d: invalid sentence order %q", cuePosition, strings.TrimSpace(lines[0]))
		}
		if _, exists := seenOrders[order]; exists {
			return nil, fmt.Errorf("cue position %d: duplicate sentence order %d", cuePosition, order)
		}
		seenOrders[order] = struct{}{}
		times := strings.Split(strings.TrimSpace(lines[1]), "-->")
		if len(times) != 2 {
			return nil, fmt.Errorf("cue order %d: invalid timestamp line %q", order, strings.TrimSpace(lines[1]))
		}
		start, err := parseTimestamp(strings.TrimSpace(times[0]))
		if err != nil {
			return nil, fmt.Errorf("cue order %d: invalid start timestamp: %v", order, err)
		}
		endFields := strings.Fields(strings.TrimSpace(times[1]))
		if len(endFields) == 0 {
			return nil, fmt.Errorf("cue order %d: missing end timestamp", order)
		}
		end, err := parseTimestamp(endFields[0])
		if err != nil {
			return nil, fmt.Errorf("cue order %d: invalid end timestamp: %v", order, err)
		}
		if end < start {
			return nil, fmt.Errorf("cue order %d: end timestamp is before start timestamp", order)
		}
		content := strings.TrimSpace(strings.Join(lines[2:], "\n"))
		if content == "" {
			return nil, fmt.Errorf("cue order %d: content is empty", order)
		}
		result = append(result, entity.CaptionInput{SentenceOrder: order, StartTimeMS: start, EndTimeMS: end, Content: content})
	}
	return result, nil
}

func parseTimestamp(raw string) (int64, error) {
	parts := strings.Split(raw, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("expected HH:MM:SS,mmm, got %q", raw)
	}
	secondsAndMillis := strings.FieldsFunc(parts[2], func(r rune) bool { return r == ',' || r == '.' })
	if len(secondsAndMillis) != 2 || len(secondsAndMillis[1]) != 3 {
		return 0, fmt.Errorf("milliseconds must contain exactly 3 digits in %q", raw)
	}
	hours, errH := strconv.ParseInt(parts[0], 10, 64)
	minutes, errM := strconv.ParseInt(parts[1], 10, 64)
	seconds, errS := strconv.ParseInt(secondsAndMillis[0], 10, 64)
	millis, errMS := strconv.ParseInt(secondsAndMillis[1], 10, 64)
	if errH != nil || errM != nil || errS != nil || errMS != nil || hours < 0 || minutes < 0 || minutes > 59 || seconds < 0 || seconds > 59 {
		return 0, fmt.Errorf("timestamp value is out of range in %q", raw)
	}
	return ((hours*60+minutes)*60+seconds)*1000 + millis, nil
}
