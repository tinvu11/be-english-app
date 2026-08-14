package v1

import (
	"errors"
	"net/http"
	"strconv"

	_ "github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary Get AI-generated video learning content
// @Description Returns shared English quizzes plus a summary and reusable dictionary vocabulary in the current user's native language. Missing parts are generated and persisted.
// @ID user-video-learning-content
// @Tags user-videos
// @Produce json
// @Param videoId path int true "Video ID"
// @Success 200 {object} entity.VideoLearningContent
// @Failure 400,401,404,422,502,503,500 {object} response.Error
// @Security BearerAuth
// @Router /videos/{videoId}/learning-content [get]
func (r *V1) getVideoLearningContent(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	videoID, err := strconv.ParseInt(ctx.Params("videoId"), 10, 64)
	if !ok || userID == "" {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	if err != nil || videoID <= 0 {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video id")
	}
	result, err := r.learning.GetForUser(ctx.UserContext(), userID, videoID)
	if err == nil {
		return ctx.JSON(result)
	}
	switch {
	case errors.Is(err, entity.ErrVideoNotFound):
		return errorResponse(ctx, http.StatusNotFound, "video not found")
	case errors.Is(err, entity.ErrNativeLanguageRequired):
		return errorResponse(ctx, http.StatusUnprocessableEntity, "native language is required")
	case errors.Is(err, entity.ErrCaptionTranslationEmpty):
		return errorResponse(ctx, http.StatusUnprocessableEntity, "video has no captions")
	case errors.Is(err, entity.ErrInvalidLanguage), errors.Is(err, entity.ErrInvalidLearningContent):
		return errorResponse(ctx, http.StatusUnprocessableEntity, "invalid learning content")
	case errors.Is(err, entity.ErrTranslationUnavailable):
		return errorResponse(ctx, http.StatusServiceUnavailable, "AI generation is not configured")
	case errors.Is(err, entity.ErrTranslationRateLimited):
		return errorResponse(ctx, http.StatusTooManyRequests, "AI provider rate limit reached")
	case errors.Is(err, entity.ErrTranslationFailed), errors.Is(err, entity.ErrLearningContentFailed):
		return errorResponse(ctx, http.StatusBadGateway, "AI generation failed")
	default:
		r.l.Error(err, "restapi - v1 - learning content")
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}
