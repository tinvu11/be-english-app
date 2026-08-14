package admin

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	adminrequest "github.com/evrone/go-clean-template/internal/controller/restapi/v1/admin/request"
	_ "github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary Preview YouTube video metadata
// @Description Returns YouTube metadata without saving the video or downloading audio
// @Tags admin-videos
// @Accept json
// @Produce json
// @Param request body adminrequest.YouTubePreview true "YouTube URL or ID"
// @Success 200 {object} entity.YouTubeVideoPreview
// @Failure 400,401,403,502,504,500 {object} response.Error
// @Security BearerAuth
// @Router /admin/videos/youtube-preview [post]
func (ctrl *controller) previewYouTubeVideo(ctx *fiber.Ctx) error {
	var body adminrequest.YouTubePreview
	if err := ctx.BodyParser(&body); err != nil || ctrl.validate.Struct(body) != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid YouTube URL")
	}
	item, err := ctrl.videos.PreviewYouTubeVideo(ctx.UserContext(), body.YouTubeURL)
	if err != nil {
		return ctrl.videoError(ctx, err)
	}
	return ctx.JSON(item)
}

// @Summary List videos
// @Description Returns videos with language, level, channel, and translated topics
// @Tags admin-videos
// @Produce json
// @Param language_id query int false "Language ID"
// @Param level_id query int false "Level ID"
// @Param channel_id query int false "Channel ID"
// @Param status query string false "Status" Enums(draft,published,archived)
// @Param search query string false "Search title"
// @Param limit query int false "Page size (1-100)" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} entity.VideoList
// @Failure 400,401,403,500 {object} response.Error
// @Security BearerAuth
// @Router /admin/videos [get]
func (ctrl *controller) listVideos(ctx *fiber.Ctx) error {
	filter := entity.VideoFilter{Status: ctx.Query("status"), Search: ctx.Query("search"), Limit: 20}
	var err error
	if filter.LanguageID, err = optionalQueryID(ctx, "language_id"); err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid language_id")
	}
	if filter.LevelID, err = optionalQueryID(ctx, "level_id"); err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid level_id")
	}
	if filter.ChannelID, err = optionalQueryID(ctx, "channel_id"); err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid channel_id")
	}
	if raw := ctx.Query("limit"); raw != "" {
		filter.Limit, err = strconv.Atoi(raw)
		if err != nil {
			return errorResponse(ctx, http.StatusBadRequest, "invalid limit")
		}
	}
	if raw := ctx.Query("offset"); raw != "" {
		filter.Offset, err = strconv.Atoi(raw)
		if err != nil {
			return errorResponse(ctx, http.StatusBadRequest, "invalid offset")
		}
	}
	result, err := ctrl.videos.ListVideos(ctx.UserContext(), filter)
	if err != nil {
		return ctrl.videoError(ctx, err)
	}
	return ctx.JSON(result)
}

// @Summary Get video details
// @Tags admin-videos
// @Produce json
// @Param id path int true "Video ID"
// @Success 200 {object} entity.Video
// @Failure 400,401,403,404,500 {object} response.Error
// @Security BearerAuth
// @Router /admin/videos/{id} [get]
func (ctrl *controller) getVideo(ctx *fiber.Ctx) error {
	id, err := positiveInt64(ctx.Params("id"))
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video id")
	}
	item, err := ctrl.videos.GetVideo(ctx.UserContext(), id)
	if err != nil {
		return ctrl.videoError(ctx, err)
	}
	return ctx.JSON(item)
}

// @Summary Create video
// @Tags admin-videos
// @Accept json
// @Produce json
// @Param request body adminrequest.SaveVideo true "Video"
// @Success 201 {object} entity.Video
// @Failure 400,401,403,409,500 {object} response.Error
// @Security BearerAuth
// @Router /admin/videos [post]
func (ctrl *controller) createVideo(ctx *fiber.Ctx) error {
	body, err := ctrl.parseVideo(ctx)
	if err != nil {
		return err
	}
	item, err := ctrl.videos.CreateVideo(ctx.UserContext(), videoFromRequest(body))
	if err != nil {
		return ctrl.videoError(ctx, err)
	}
	return ctx.Status(http.StatusCreated).JSON(item)
}

// @Summary Update video
// @Tags admin-videos
// @Accept json
// @Produce json
// @Param id path int true "Video ID"
// @Param request body adminrequest.SaveVideo true "Video"
// @Success 200 {object} entity.Video
// @Failure 400,401,403,404,409,500 {object} response.Error
// @Security BearerAuth
// @Router /admin/videos/{id} [put]
func (ctrl *controller) updateVideo(ctx *fiber.Ctx) error {
	id, err := positiveInt64(ctx.Params("id"))
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video id")
	}
	body, err := ctrl.parseVideo(ctx)
	if err != nil {
		return err
	}
	item, err := ctrl.videos.UpdateVideo(ctx.UserContext(), id, videoFromRequest(body))
	if err != nil {
		return ctrl.videoError(ctx, err)
	}
	return ctx.JSON(item)
}

// @Summary Transition video status
// @Description Only draft to published and published to archived transitions are allowed
// @Tags admin-videos
// @Accept json
// @Produce json
// @Param id path int true "Video ID"
// @Param request body adminrequest.VideoStatus true "Next status"
// @Success 200 {object} entity.Video
// @Failure 400,401,403,404,409,500 {object} response.Error
// @Security BearerAuth
// @Router /admin/videos/{id}/status [patch]
func (ctrl *controller) setVideoStatus(ctx *fiber.Ctx) error {
	id, err := positiveInt64(ctx.Params("id"))
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video id")
	}
	var body adminrequest.VideoStatus
	if err = ctx.BodyParser(&body); err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}
	if err = ctrl.validate.Struct(body); err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video status")
	}
	item, err := ctrl.videos.TransitionVideoStatus(ctx.UserContext(), id, body.Status)
	if err != nil {
		return ctrl.videoError(ctx, err)
	}
	return ctx.JSON(item)
}

// @Summary Delete video
// @Tags admin-videos
// @Param id path int true "Video ID"
// @Success 204
// @Failure 400,401,403,404,500 {object} response.Error
// @Security BearerAuth
// @Router /admin/videos/{id} [delete]
func (ctrl *controller) deleteVideo(ctx *fiber.Ctx) error {
	id, err := positiveInt64(ctx.Params("id"))
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video id")
	}
	if err = ctrl.videos.DeleteVideo(ctx.UserContext(), id); err != nil {
		return ctrl.videoError(ctx, err)
	}
	return ctx.SendStatus(http.StatusNoContent)
}

func (ctrl *controller) parseVideo(ctx *fiber.Ctx) (adminrequest.SaveVideo, error) {
	var body adminrequest.SaveVideo
	if err := ctx.BodyParser(&body); err != nil {
		return body, errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}
	if err := ctrl.validate.Struct(body); err != nil {
		return body, errorResponse(ctx, http.StatusBadRequest, "invalid video")
	}
	return body, nil
}

func videoFromRequest(body adminrequest.SaveVideo) entity.VideoInput {
	return entity.VideoInput{Title: body.Title, YouTubeID: body.YouTubeID, ThumbnailURL: body.ThumbnailURL,
		DurationSeconds: body.DurationSeconds, Status: body.Status, LanguageID: body.LanguageID,
		LevelID: body.LevelID, ChannelID: body.ChannelID, TopicIDs: body.TopicIDs}
}

func optionalQueryID(ctx *fiber.Ctx, name string) (*int, error) {
	raw := ctx.Query(name)
	if raw == "" {
		return nil, nil
	}
	id, err := positiveID(raw)
	return &id, err
}

func positiveInt64(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, entity.ErrInvalidVideo
	}
	return id, nil
}

func (ctrl *controller) videoError(ctx *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, entity.ErrInvalidVideo), errors.Is(err, entity.ErrInvalidReference):
		return errorResponse(ctx, http.StatusBadRequest, "invalid video or related reference")
	case errors.Is(err, entity.ErrVideoNotFound):
		return errorResponse(ctx, http.StatusNotFound, "video not found")
	case errors.Is(err, entity.ErrVideoExists):
		return errorResponse(ctx, http.StatusConflict, "YouTube video already exists")
	case errors.Is(err, entity.ErrInvalidVideoTransition):
		return errorResponse(ctx, http.StatusConflict, "invalid video status transition")
	case errors.Is(err, entity.ErrConcurrentVideoUpdate):
		return errorResponse(ctx, http.StatusConflict, "video status changed concurrently")
	case errors.Is(err, entity.ErrYouTubeProviderFailed):
		if errors.Is(err, context.DeadlineExceeded) {
			return errorResponse(ctx, http.StatusGatewayTimeout, "YouTube metadata request timed out")
		}
		return errorResponse(ctx, http.StatusBadGateway, "YouTube metadata provider failed")
	default:
		ctrl.log.Error(err, "restapi - v1 - admin - video")
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}
