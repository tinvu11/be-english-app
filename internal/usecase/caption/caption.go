package caption

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/gateway"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/internal/usecase"
)

const maxImportCaptions = 10000
const maxCaptionPageSize = 200

type UseCase struct {
	repo           repo.CaptionRepo
	videos         repo.VideoRepo
	languages      repo.LanguageRepo
	users          repo.UserRepo
	subtitles      gateway.YouTubeSubtitleProvider
	translator     gateway.CaptionTranslator
	translateBatch int
}

func New(repository repo.CaptionRepo, videos repo.VideoRepo, languages repo.LanguageRepo, users repo.UserRepo, subtitles gateway.YouTubeSubtitleProvider, translator gateway.CaptionTranslator, translateBatch int) usecase.Caption {
	if translateBatch <= 0 {
		translateBatch = 50
	}
	return newTraced(&UseCase{repo: repository, videos: videos, languages: languages, users: users, subtitles: subtitles,
		translator: translator, translateBatch: translateBatch})
}

func (uc *UseCase) GetOriginalCaptions(ctx context.Context, videoID int64) (entity.VideoCaptions, error) {
	if videoID <= 0 {
		return entity.VideoCaptions{}, entity.ErrInvalidCaption
	}
	video, err := uc.publishedVideo(ctx, videoID)
	if err != nil {
		return entity.VideoCaptions{}, err
	}
	list, err := uc.allCaptions(ctx, videoID)
	if err != nil {
		return entity.VideoCaptions{}, err
	}
	result := entity.VideoCaptions{VideoID: videoID, LanguageID: video.Language.ID, LanguageCode: video.Language.Code,
		Items: make([]entity.VideoCaptionItem, 0, len(list.Items)), Total: list.Total}
	for _, item := range list.Items {
		result.Items = append(result.Items, captionItem(item, item.Content))
	}
	return result, nil
}

func (uc *UseCase) GetTranslatedCaptions(ctx context.Context, userID string, videoID int64) (entity.VideoCaptions, error) {
	if strings.TrimSpace(userID) == "" || videoID <= 0 {
		return entity.VideoCaptions{}, entity.ErrInvalidCaption
	}
	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return entity.VideoCaptions{}, err
	}
	if user.NativeLanguageID == nil || *user.NativeLanguageID <= 0 {
		return entity.VideoCaptions{}, entity.ErrNativeLanguageRequired
	}
	video, err := uc.publishedVideo(ctx, videoID)
	if err != nil {
		return entity.VideoCaptions{}, err
	}
	target, err := uc.languages.GetLanguage(ctx, *user.NativeLanguageID)
	if err != nil {
		return entity.VideoCaptions{}, err
	}
	if !target.IsActive {
		return entity.VideoCaptions{}, entity.ErrInvalidLanguage
	}
	sameLanguage := strings.EqualFold(video.Language.Code, target.Code)
	if !sameLanguage {
		if _, err = uc.TranslateCaptions(ctx, videoID, target.ID, entity.CaptionTranslationMissingOnly); err != nil {
			return entity.VideoCaptions{}, err
		}
	}
	list, err := uc.allCaptions(ctx, videoID)
	if err != nil {
		return entity.VideoCaptions{}, err
	}
	result := entity.VideoCaptions{VideoID: videoID, LanguageID: target.ID, LanguageCode: target.Code,
		Items: make([]entity.VideoCaptionItem, 0, len(list.Items)), Total: list.Total}
	for _, item := range list.Items {
		text := item.Content
		if !sameLanguage {
			text = ""
			for _, translation := range item.Translations {
				if translation.LanguageID == target.ID {
					text = translation.Text
					break
				}
			}
			if text == "" {
				return entity.VideoCaptions{}, entity.ErrInvalidTranslation
			}
		}
		result.Items = append(result.Items, captionItem(item, text))
	}
	return result, nil
}

func (uc *UseCase) publishedVideo(ctx context.Context, videoID int64) (entity.Video, error) {
	video, err := uc.videos.GetVideo(ctx, videoID)
	if err != nil {
		return entity.Video{}, err
	}
	if video.Status != entity.VideoStatusPublished {
		return entity.Video{}, entity.ErrVideoNotFound
	}
	return video, nil
}

func (uc *UseCase) allCaptions(ctx context.Context, videoID int64) (entity.CaptionList, error) {
	list, err := uc.repo.ListCaptions(ctx, videoID, entity.CaptionFilter{Limit: maxImportCaptions, Offset: 0})
	if err != nil {
		return entity.CaptionList{}, err
	}
	if list.Total == 0 {
		return entity.CaptionList{}, entity.ErrCaptionTranslationEmpty
	}
	return list, nil
}

func captionItem(item entity.Caption, text string) entity.VideoCaptionItem {
	return entity.VideoCaptionItem{ID: item.ID, SentenceOrder: item.SentenceOrder,
		StartTimeMS: item.StartTimeMS, EndTimeMS: item.EndTimeMS, Text: text}
}

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

func (uc *UseCase) ListYouTubeSubtitleTracks(ctx context.Context, videoID int64) (entity.YouTubeSubtitleTracks, error) {
	if videoID <= 0 {
		return entity.YouTubeSubtitleTracks{}, entity.ErrInvalidVideo
	}
	video, err := uc.videos.GetVideo(ctx, videoID)
	if err != nil {
		return entity.YouTubeSubtitleTracks{}, err
	}
	tracks, err := uc.subtitles.ListManualSubtitles(ctx, video.YouTubeID)
	if err != nil {
		return entity.YouTubeSubtitleTracks{}, err
	}
	return entity.YouTubeSubtitleTracks{Tracks: tracks}, nil
}

func (uc *UseCase) ImportFromYouTube(ctx context.Context, videoID int64, languageCode, mode string) (entity.YouTubeCaptionImportResult, error) {
	languageCode = strings.TrimSpace(languageCode)
	mode = strings.ToLower(strings.TrimSpace(mode))
	if videoID <= 0 || languageCode == "" || len(languageCode) > 35 ||
		(mode != entity.CaptionImportFailIfExists && mode != entity.CaptionImportReplaceAll) {
		return entity.YouTubeCaptionImportResult{}, entity.ErrInvalidCaption
	}
	video, err := uc.videos.GetVideo(ctx, videoID)
	if err != nil {
		return entity.YouTubeCaptionImportResult{}, err
	}
	data, err := uc.subtitles.DownloadManualSubtitle(ctx, video.YouTubeID, languageCode)
	if err != nil {
		return entity.YouTubeCaptionImportResult{}, err
	}
	inputs, err := parseWebVTT(data)
	if err != nil || len(inputs) == 0 || len(inputs) > maxImportCaptions {
		return entity.YouTubeCaptionImportResult{}, fmt.Errorf("%w: %v", entity.ErrInvalidWebVTT, err)
	}
	if mode == entity.CaptionImportReplaceAll {
		_, err = uc.repo.ReplaceCaptions(ctx, videoID, inputs)
	} else {
		_, err = uc.repo.ImportCaptionsIfEmpty(ctx, videoID, inputs)
	}
	if err != nil {
		return entity.YouTubeCaptionImportResult{}, err
	}
	source := "youtube_manual"
	if tracks, listErr := uc.subtitles.ListManualSubtitles(ctx, video.YouTubeID); listErr == nil {
		for _, track := range tracks {
			if track.LanguageCode == languageCode && track.IsAutomatic {
				source = "youtube_auto"
				break
			}
		}
	}
	return entity.YouTubeCaptionImportResult{VideoID: videoID, LanguageCode: languageCode,
		Source: source, ImportedCount: len(inputs)}, nil
}

func (uc *UseCase) TranslateCaptions(ctx context.Context, videoID int64, targetLanguageID int, mode string) (entity.CaptionTranslationResult, error) {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if videoID <= 0 || targetLanguageID <= 0 || (mode != entity.CaptionTranslationMissingOnly && mode != entity.CaptionTranslationReplace) {
		return entity.CaptionTranslationResult{}, entity.ErrInvalidCaption
	}
	video, err := uc.videos.GetVideo(ctx, videoID)
	if err != nil {
		return entity.CaptionTranslationResult{}, err
	}
	target, err := uc.languages.GetLanguage(ctx, targetLanguageID)
	if err != nil {
		return entity.CaptionTranslationResult{}, err
	}
	if !target.IsActive || strings.EqualFold(video.Language.Code, target.Code) {
		return entity.CaptionTranslationResult{}, entity.ErrInvalidLanguage
	}
	items, total, err := uc.repo.ListCaptionsForTranslation(ctx, videoID, targetLanguageID, mode == entity.CaptionTranslationMissingOnly)
	if err != nil {
		return entity.CaptionTranslationResult{}, err
	}
	if total == 0 {
		return entity.CaptionTranslationResult{}, entity.ErrCaptionTranslationEmpty
	}
	result := entity.CaptionTranslationResult{VideoID: videoID, SourceLanguageCode: video.Language.Code,
		TargetLanguageCode: target.Code, SkippedCount: total - len(items)}
	if len(items) == 0 {
		return result, nil
	}
	translations := make([]entity.CaptionTranslationUpsert, 0, len(items))
	for start := 0; start < len(items); start += uc.translateBatch {
		end := min(start+uc.translateBatch, len(items))
		batch := items[start:end]
		requestItems := make([]entity.CaptionText, len(batch))
		for index, caption := range batch {
			requestItems[index] = entity.CaptionText{CaptionID: caption.ID, Order: caption.SentenceOrder, Text: caption.Content}
		}
		translated, translateErr := uc.translator.TranslateCaptions(ctx, entity.CaptionTranslationRequest{
			SourceLanguage: video.Language.Code, TargetLanguage: target.Code, Items: requestItems,
		})
		if translateErr != nil {
			return entity.CaptionTranslationResult{}, translateErr
		}
		validated, validateErr := validateTranslationBatch(requestItems, translated)
		if validateErr != nil {
			return entity.CaptionTranslationResult{}, validateErr
		}
		translations = append(translations, validated...)
	}
	if err = uc.repo.UpsertTranslations(ctx, videoID, targetLanguageID, translations); err != nil {
		return entity.CaptionTranslationResult{}, err
	}
	result.TranslatedCount = len(translations)
	return result, nil
}

func validateTranslationBatch(source []entity.CaptionText, translated []entity.TranslatedCaption) ([]entity.CaptionTranslationUpsert, error) {
	if len(source) != len(translated) {
		return nil, fmt.Errorf("%w: item count mismatch expected=%d actual=%d", entity.ErrInvalidTranslation, len(source), len(translated))
	}
	expected := make(map[int64]int, len(source))
	for _, item := range source {
		expected[item.CaptionID] = item.Order
	}
	result := make([]entity.CaptionTranslationUpsert, 0, len(translated))
	seen := make(map[int64]struct{}, len(translated))
	for _, item := range translated {
		order, exists := expected[item.CaptionID]
		item.Text = strings.TrimSpace(item.Text)
		if !exists {
			return nil, fmt.Errorf("%w: unexpected caption_id=%d", entity.ErrInvalidTranslation, item.CaptionID)
		}
		if order != item.Order {
			return nil, fmt.Errorf("%w: caption_id=%d order mismatch expected=%d actual=%d", entity.ErrInvalidTranslation, item.CaptionID, order, item.Order)
		}
		if item.Text == "" {
			return nil, fmt.Errorf("%w: empty text caption_id=%d order=%d", entity.ErrInvalidTranslation, item.CaptionID, item.Order)
		}
		if _, duplicate := seen[item.CaptionID]; duplicate {
			return nil, fmt.Errorf("%w: duplicate caption_id=%d", entity.ErrInvalidTranslation, item.CaptionID)
		}
		seen[item.CaptionID] = struct{}{}
		result = append(result, entity.CaptionTranslationUpsert{CaptionID: item.CaptionID, Text: item.Text})
	}
	return result, nil
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
