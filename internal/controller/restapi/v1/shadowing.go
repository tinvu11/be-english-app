package v1

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

const maxShadowingUploadBytes = 5 << 20

// @Summary Assess a shadowing attempt
// @Description Upload WAV PCM 16 kHz or OGG/Opus audio and assess it against the selected caption.
// @ID assess-shadowing
// @Tags shadowing
// @Accept multipart/form-data
// @Produce json
// @Param videoId path int true "Video ID"
// @Param captionId path int true "Caption ID"
// @Param audio formData file true "Recorded audio (max 5 MB)"
// @Param locale formData string false "Azure BCP-47 locale; defaults from video language"
// @Success 201 {object} entity.ShadowingAttempt
// @Failure 400,401,404,413,422,502,503,500 {object} map[string]string
// @Failure 429 {object} response.QuotaError
// @Security BearerAuth
// @Router /videos/{videoId}/shadowing-attempts/{captionId} [post]
func (r *V1) assessShadowing(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	videoID, videoErr := strconv.ParseInt(ctx.Params("videoId"), 10, 64)
	captionID, captionErr := strconv.ParseInt(ctx.Params("captionId"), 10, 64)
	if !ok || userID == "" {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	if videoErr != nil || captionErr != nil || videoID <= 0 || captionID <= 0 {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video or caption id")
	}
	header, err := ctx.FormFile("audio")
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "audio file is required")
	}
	if header.Size <= 0 || header.Size > maxShadowingUploadBytes {
		return errorResponse(ctx, http.StatusRequestEntityTooLarge, "audio file must not exceed 5 MB")
	}
	file, err := header.Open()
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "cannot read audio file")
	}
	defer file.Close()
	audio, err := io.ReadAll(io.LimitReader(file, maxShadowingUploadBytes+1))
	if err != nil || len(audio) > maxShadowingUploadBytes {
		return errorResponse(ctx, http.StatusRequestEntityTooLarge, "audio file must not exceed 5 MB")
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(audio)
	}
	attempt, err := r.shadowing.Assess(ctx.UserContext(), userID, videoID, captionID, audio, contentType, strings.TrimSpace(ctx.FormValue("locale")))
	if err != nil {
		return r.shadowingError(ctx, err)
	}
	return ctx.Status(http.StatusCreated).JSON(attempt)
}

// @Summary List shadowing attempts
// @Description Return the current user's attempts for a video, optionally filtered by caption.
// @ID list-shadowing-attempts
// @Tags shadowing
// @Produce json
// @Param videoId path int true "Video ID"
// @Param caption_id query int false "Caption ID"
// @Success 200 {object} entity.ShadowingAttemptList
// @Failure 400,401,500 {object} map[string]string
// @Security BearerAuth
// @Router /videos/{videoId}/shadowing-attempts [get]
func (r *V1) listShadowingAttempts(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	videoID, err := strconv.ParseInt(ctx.Params("videoId"), 10, 64)
	if !ok || userID == "" {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	if err != nil || videoID <= 0 {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video id")
	}
	var captionID int64
	if value := strings.TrimSpace(ctx.Query("caption_id")); value != "" {
		captionID, err = strconv.ParseInt(value, 10, 64)
		if err != nil || captionID <= 0 {
			return errorResponse(ctx, http.StatusBadRequest, "invalid caption id")
		}
	}
	result, err := r.shadowing.ListAttempts(ctx.UserContext(), userID, videoID, captionID)
	if err != nil {
		return r.shadowingError(ctx, err)
	}
	return ctx.JSON(result)
}

func (r *V1) shadowingError(ctx *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, entity.ErrQuotaExceeded):
		return quotaErrorResponse(ctx, err)
	case errors.Is(err, entity.ErrInvalidShadowingAudio), errors.Is(err, entity.ErrInvalidLanguage):
		return errorResponse(ctx, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, entity.ErrCaptionNotFound):
		return errorResponse(ctx, http.StatusNotFound, "caption not found for video")
	case errors.Is(err, entity.ErrPronunciationUnavailable):
		return errorResponse(ctx, http.StatusServiceUnavailable, "pronunciation assessment is not configured")
	case errors.Is(err, entity.ErrPronunciationFailed), errors.Is(err, entity.ErrInvalidPronunciation):
		return errorResponse(ctx, http.StatusBadGateway, "pronunciation assessment failed")
	default:
		r.l.Error(err, "restapi - v1 - shadowing")
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}
