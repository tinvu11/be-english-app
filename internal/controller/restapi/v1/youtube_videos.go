package v1

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/request"
	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary Preview a YouTube video
// @Description Returns YouTube metadata without saving the video or downloading audio
// @ID user-preview-youtube-video
// @Tags user-videos
// @Accept json
// @Produce json
// @Param request body request.YouTubePreview true "YouTube URL or ID"
// @Success 200 {object} entity.YouTubeVideoPreview
// @Failure 400,401,502 {object} map[string]string
// @Security BearerAuth
// @Router /videos/youtube-preview [post]
func (r *V1) previewUserYouTubeVideo(ctx *fiber.Ctx) error {
	var body request.YouTubePreview
	if err := ctx.BodyParser(&body); err != nil || r.v.Struct(body) != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid YouTube URL")
	}
	preview, err := r.videos.PreviewYouTubeVideo(ctx.UserContext(), body.YouTubeURL)
	if err != nil {
		return r.youtubeVideoError(ctx, err, "preview")
	}
	return ctx.JSON(preview)
}

// @Summary Get original video captions
// @Description Returns original captions immediately without requesting a translation
// @ID user-video-original-captions
// @Tags user-videos
// @Produce json
// @Param videoId path int true "Video ID"
// @Success 200 {object} response.OriginalVideoCaptions
// @Failure 400,401,404 {object} map[string]string
// @Security BearerAuth
// @Router /videos/{videoId}/captions [get]
func (r *V1) getOriginalVideoCaptions(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok || userID == "" {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	videoID, err := strconv.ParseInt(ctx.Params("videoId"), 10, 64)
	if err != nil || videoID <= 0 {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video id")
	}
	result, err := r.captions.GetOriginalCaptions(ctx.UserContext(), videoID)
	if err != nil {
		return r.userCaptionError(ctx, err)
	}
	state, err := r.u.GetVideoState(ctx.UserContext(), userID, videoID)
	if err != nil {
		return r.userCaptionError(ctx, err)
	}
	completed, err := r.u.ListCompletedDictations(ctx.UserContext(), userID, videoID)
	if err != nil {
		return r.userCaptionError(ctx, err)
	}
	return ctx.JSON(buildOriginalVideoCaptions(result, state, completed))
}

func buildOriginalVideoCaptions(captions entity.VideoCaptions, state entity.VideoState, completed entity.DictationProgressList) response.OriginalVideoCaptions {
	completedIDs := make(map[int64]struct{}, len(completed.Items))
	for _, item := range completed.Items {
		completedIDs[item.CaptionID] = struct{}{}
	}
	result := response.OriginalVideoCaptions{
		VideoID: captions.VideoID, LanguageID: captions.LanguageID, LanguageCode: captions.LanguageCode,
		VideoState: state, Items: make([]response.VideoCaptionItem, 0, len(captions.Items)), Total: captions.Total,
	}
	for _, item := range captions.Items {
		_, isCompleted := completedIDs[item.ID]
		result.Items = append(result.Items, response.VideoCaptionItem{
			ID: item.ID, SentenceOrder: item.SentenceOrder, StartTimeMS: item.StartTimeMS,
			EndTimeMS: item.EndTimeMS, Text: item.Text, DictationCompleted: isCompleted,
		})
	}
	return result
}

// @Summary Get translated video captions
// @Description Returns captions translated to the user's native language; missing translations are generated and persisted before returning
// @ID user-video-translated-captions
// @Tags user-videos
// @Produce json
// @Param videoId path int true "Video ID"
// @Success 200 {object} entity.VideoCaptions
// @Failure 400,401,404,409,502 {object} map[string]string
// @Failure 429 {object} response.QuotaError
// @Security BearerAuth
// @Router /videos/{videoId}/captions/translation [get]
func (r *V1) getTranslatedVideoCaptions(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	videoID, err := strconv.ParseInt(ctx.Params("videoId"), 10, 64)
	if !ok || userID == "" {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	if err != nil || videoID <= 0 {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video id")
	}
	result, err := r.captions.GetTranslatedCaptions(ctx.UserContext(), userID, videoID)
	if err != nil {
		return r.userCaptionError(ctx, err)
	}
	return ctx.JSON(result)
}

func (r *V1) userCaptionError(ctx *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, entity.ErrNativeLanguageRequired):
		return errorResponse(ctx, http.StatusConflict, "native language must be selected first")
	case errors.Is(err, entity.ErrVideoNotFound), errors.Is(err, entity.ErrCaptionTranslationEmpty):
		return errorResponse(ctx, http.StatusNotFound, err.Error())
	case errors.Is(err, entity.ErrInvalidCaption), errors.Is(err, entity.ErrInvalidLanguage):
		return errorResponse(ctx, http.StatusBadRequest, err.Error())
	case errors.Is(err, entity.ErrTranslationRateLimited):
		return errorResponse(ctx, http.StatusTooManyRequests, "translation rate limit reached")
	case errors.Is(err, entity.ErrTranslationUnavailable), errors.Is(err, entity.ErrTranslationFailed),
		errors.Is(err, entity.ErrInvalidTranslation):
		return errorResponse(ctx, http.StatusBadGateway, "caption translation failed")
	default:
		r.l.Error(err, "restapi - v1 - user captions")
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}

// @Summary Add a YouTube video for the current user
// @Description Reuses an existing video and captions when available; otherwise downloads normalized audio and transcribes it with Groq Whisper. audioLanguageCode is an optional spoken-language hint.
// @ID user-import-youtube-video
// @Tags user-videos
// @Accept json
// @Produce json
// @Param request body request.ImportYouTubeVideo true "YouTube video import"
// @Success 200 {object} entity.UserYouTubeVideoResult
// @Failure 400,401,404,409,429,502 {object} map[string]string
// @Security BearerAuth
// @Router /videos/import-youtube [post]
func (r *V1) importUserYouTubeVideo(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	var body request.ImportYouTubeVideo
	if !ok || userID == "" {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	if err := ctx.BodyParser(&body); err != nil || r.v.Struct(body) != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid YouTube video import")
	}
	video, reused, err := r.videos.AddUserYouTubeVideo(ctx.UserContext(), userID, body.YouTubeURL)
	if err != nil {
		return r.youtubeVideoError(ctx, err, "add")
	}
	result := entity.UserYouTubeVideoResult{Video: video, Reused: reused}
	if video.CaptionAvailability.CaptionCount > 0 {
		return ctx.JSON(result)
	}
	imported, importErr := r.captions.ImportFromYouTube(ctx.UserContext(), video.ID, body.AudioLanguageCode, entity.CaptionImportFailIfExists)
	if importErr != nil && !errors.Is(importErr, entity.ErrCaptionExists) {
		return r.youtubeVideoError(ctx, importErr, "import captions")
	}
	result.Video, err = r.videos.GetVideo(ctx.UserContext(), video.ID)
	if err != nil {
		return r.youtubeVideoError(ctx, err, "reload")
	}
	if importErr == nil {
		result.CaptionImported = true
		result.CaptionSource = imported.Source
	}
	return ctx.JSON(result)
}

func (r *V1) youtubeVideoError(ctx *fiber.Ctx, err error, operation string) error {
	switch {
	case errors.Is(err, entity.ErrQuotaExceeded):
		return quotaErrorResponse(ctx, err)
	case errors.Is(err, entity.ErrInvalidVideo), errors.Is(err, entity.ErrInvalidCaption):
		return errorResponse(ctx, http.StatusBadRequest, err.Error())
	case errors.Is(err, entity.ErrTargetLanguageRequired):
		return errorResponse(ctx, http.StatusConflict, "target language must be selected first")
	case errors.Is(err, entity.ErrYouTubeProviderFailed), errors.Is(err, entity.ErrAudioDownloadFailed):
		return errorResponse(ctx, http.StatusBadGateway, "YouTube request failed")
	case errors.Is(err, entity.ErrTranscriptionFailed), errors.Is(err, entity.ErrInvalidTranscription):
		return errorResponse(ctx, http.StatusBadGateway, "audio transcription failed")
	default:
		r.l.Error(err, "restapi - v1 - YouTube video - "+operation)
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}
