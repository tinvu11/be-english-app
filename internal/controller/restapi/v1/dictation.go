package v1

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary Complete a dictation sentence
// @Description Save a caption as completed by the current user. Repeating the request refreshes completed_at.
// @ID complete-dictation
// @Tags dictation
// @Produce json
// @Param videoId path int true "Video ID"
// @Param captionId path int true "Caption ID"
// @Success 200 {object} entity.DictationProgress
// @Failure 400,401,404,500 {object} map[string]string
// @Security BearerAuth
// @Router /videos/{videoId}/dictation-progress/{captionId} [put]
func (r *V1) completeDictation(ctx *fiber.Ctx) error {
	userID, videoID, captionID, err := dictationParams(ctx, true)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video or caption id")
	}
	item, err := r.u.CompleteDictation(ctx.UserContext(), userID, videoID, captionID)
	if err != nil {
		return r.dictationError(ctx, err, "complete")
	}
	return ctx.Status(http.StatusOK).JSON(item)
}

// @Summary Get completed dictation sentences
// @Description Return all captions completed by the current user for a video
// @ID list-completed-dictations
// @Tags dictation
// @Produce json
// @Param videoId path int true "Video ID"
// @Success 200 {object} entity.DictationProgressList
// @Failure 400,401,500 {object} map[string]string
// @Security BearerAuth
// @Router /videos/{videoId}/dictation-progress [get]
func (r *V1) listCompletedDictations(ctx *fiber.Ctx) error {
	userID, videoID, _, err := dictationParams(ctx, false)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video id")
	}
	items, err := r.u.ListCompletedDictations(ctx.UserContext(), userID, videoID)
	if err != nil {
		return r.dictationError(ctx, err, "list")
	}
	return ctx.Status(http.StatusOK).JSON(items)
}

func dictationParams(ctx *fiber.Ctx, requireCaption bool) (string, int64, int64, error) {
	userID, ok := ctx.Locals("userID").(string)
	if !ok || userID == "" {
		return "", 0, 0, entity.ErrUserNotFound
	}
	videoID, err := strconv.ParseInt(ctx.Params("videoId"), 10, 64)
	if err != nil || videoID <= 0 {
		return "", 0, 0, entity.ErrInvalidCaption
	}
	if !requireCaption {
		return userID, videoID, 0, nil
	}
	captionID, err := strconv.ParseInt(ctx.Params("captionId"), 10, 64)
	if err != nil || captionID <= 0 {
		return "", 0, 0, entity.ErrInvalidCaption
	}
	return userID, videoID, captionID, nil
}

func (r *V1) dictationError(ctx *fiber.Ctx, err error, operation string) error {
	switch {
	case errors.Is(err, entity.ErrInvalidCaption):
		return errorResponse(ctx, http.StatusBadRequest, "invalid video or caption")
	case errors.Is(err, entity.ErrCaptionNotFound):
		return errorResponse(ctx, http.StatusNotFound, "caption not found for video")
	default:
		r.l.Error(err, "restapi - v1 - dictation - "+operation)
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}
