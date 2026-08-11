package v1

import (
	"errors"
	"net/http"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/request"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary Preview a YouTube video
// @Description Returns metadata and available caption tracks without saving the video
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

// @Summary Add a YouTube video for the current user
// @Description Reuses an existing video and captions when available; otherwise creates it using the user's target language and imports manual captions before falling back to automatic captions
// @ID user-import-youtube-video
// @Tags user-videos
// @Accept json
// @Produce json
// @Param request body request.ImportYouTubeVideo true "YouTube video import"
// @Success 200 {object} entity.UserYouTubeVideoResult
// @Failure 400,401,404,409,502 {object} map[string]string
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
	imported, importErr := r.captions.ImportFromYouTube(ctx.UserContext(), video.ID, body.CaptionLanguageCode, entity.CaptionImportFailIfExists)
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
	case errors.Is(err, entity.ErrInvalidVideo), errors.Is(err, entity.ErrInvalidCaption):
		return errorResponse(ctx, http.StatusBadRequest, err.Error())
	case errors.Is(err, entity.ErrTargetLanguageRequired):
		return errorResponse(ctx, http.StatusConflict, "target language must be selected first")
	case errors.Is(err, entity.ErrManualSubtitleNotFound):
		return errorResponse(ctx, http.StatusNotFound, "YouTube subtitle not found")
	case errors.Is(err, entity.ErrSubtitleDownloadFailed):
		return errorResponse(ctx, http.StatusBadGateway, "YouTube request failed")
	default:
		r.l.Error(err, "restapi - v1 - YouTube video - "+operation)
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}
