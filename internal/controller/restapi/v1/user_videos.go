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

// @Summary Get watch history
// @Description Return the current user's recently watched published videos
// @ID user-watch-history
// @Tags user-videos
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Videos per page (1-100)" default(20)
// @Success 200 {object} response.WatchHistory
// @Failure 400,401,500 {object} response.Error
// @Security BearerAuth
// @Router /user/watch-history [get]
func (r *V1) watchHistory(ctx *fiber.Ctx) error {
	userID, page, limit, err := userVideoRequest(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid pagination")
	}
	list, err := r.u.ListWatchHistory(ctx.UserContext(), userID, limit, (page-1)*limit)
	if err != nil {
		return r.userVideoError(ctx, err, "watch history")
	}
	return ctx.Status(http.StatusOK).JSON(buildWatchHistory(list, page, limit))
}

// @Summary Check video watch status
// @Description Check whether the current user has watched a video and return the latest position for resuming playback
// @ID user-watch-status
// @Tags user-videos
// @Produce json
// @Param videoId path int true "Video ID"
// @Success 200 {object} response.WatchStatus
// @Failure 400,401,500 {object} response.Error
// @Security BearerAuth
// @Router /user/watch-history/{videoId} [get]
func (r *V1) watchStatus(ctx *fiber.Ctx) error {
	userID, videoID, err := userVideoMutation(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video id")
	}
	video, watched, err := r.u.GetWatchHistory(ctx.UserContext(), userID, videoID)
	if err != nil {
		return r.userVideoMutationError(ctx, err, "get watch status")
	}
	return ctx.Status(http.StatusOK).JSON(buildWatchStatus(video, watched))
}

// @Summary Get saved videos
// @Description Return the current user's saved published videos
// @ID user-watch-later
// @Tags user-videos
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Videos per page (1-100)" default(20)
// @Success 200 {object} response.WatchLater
// @Failure 400,401,500 {object} response.Error
// @Security BearerAuth
// @Router /user/watch-later [get]
func (r *V1) watchLater(ctx *fiber.Ctx) error {
	userID, page, limit, err := userVideoRequest(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid pagination")
	}
	list, err := r.u.ListWatchLater(ctx.UserContext(), userID, limit, (page-1)*limit)
	if err != nil {
		return r.userVideoError(ctx, err, "watch later")
	}
	return ctx.Status(http.StatusOK).JSON(buildWatchLater(list, page, limit))
}

// @Summary Save video to watch later
// @ID user-save-watch-later
// @Tags user-videos
// @Param videoId path int true "Video ID"
// @Success 204
// @Failure 400,401,404,500 {object} response.Error
// @Security BearerAuth
// @Router /user/watch-later/{videoId} [put]
func (r *V1) saveWatchLater(ctx *fiber.Ctx) error {
	userID, videoID, err := userVideoMutation(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video id")
	}
	if err = r.u.SaveWatchLater(ctx.UserContext(), userID, videoID); err != nil {
		return r.userVideoMutationError(ctx, err, "save watch later")
	}
	return ctx.SendStatus(http.StatusNoContent)
}

// @Summary Remove video from watch later
// @ID user-remove-watch-later
// @Tags user-videos
// @Param videoId path int true "Video ID"
// @Success 204
// @Failure 400,401,500 {object} response.Error
// @Security BearerAuth
// @Router /user/watch-later/{videoId} [delete]
func (r *V1) removeWatchLater(ctx *fiber.Ctx) error {
	userID, videoID, err := userVideoMutation(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video id")
	}
	if err = r.u.RemoveWatchLater(ctx.UserContext(), userID, videoID); err != nil {
		return r.userVideoMutationError(ctx, err, "remove watch later")
	}
	return ctx.SendStatus(http.StatusNoContent)
}

// @Summary Record watch progress
// @Description Create or update a video's latest watch position and last watched time
// @ID user-upsert-watch-history
// @Tags user-videos
// @Accept json
// @Param videoId path int true "Video ID"
// @Param request body request.UpdateWatchHistory true "Watch progress"
// @Success 204
// @Failure 400,401,404,500 {object} response.Error
// @Security BearerAuth
// @Router /user/watch-history/{videoId} [put]
func (r *V1) upsertWatchHistory(ctx *fiber.Ctx) error {
	userID, videoID, err := userVideoMutation(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video id")
	}
	var body request.UpdateWatchHistory
	if err = ctx.BodyParser(&body); err != nil || r.v.Struct(body) != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid watch progress")
	}
	if err = r.u.UpsertWatchHistory(ctx.UserContext(), userID, videoID, body.LastPositionSeconds); err != nil {
		return r.userVideoMutationError(ctx, err, "upsert watch history")
	}
	return ctx.SendStatus(http.StatusNoContent)
}

// @Summary Remove video from watch history
// @ID user-remove-watch-history
// @Tags user-videos
// @Param videoId path int true "Video ID"
// @Success 204
// @Failure 400,401,500 {object} response.Error
// @Security BearerAuth
// @Router /user/watch-history/{videoId} [delete]
func (r *V1) removeWatchHistory(ctx *fiber.Ctx) error {
	userID, videoID, err := userVideoMutation(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video id")
	}
	if err = r.u.RemoveWatchHistory(ctx.UserContext(), userID, videoID); err != nil {
		return r.userVideoMutationError(ctx, err, "remove watch history")
	}
	return ctx.SendStatus(http.StatusNoContent)
}

func userVideoMutation(ctx *fiber.Ctx) (string, int64, error) {
	userID, ok := ctx.Locals("userID").(string)
	if !ok || userID == "" {
		return "", 0, entity.ErrUserNotFound
	}
	videoID, err := strconv.ParseInt(ctx.Params("videoId"), 10, 64)
	if err != nil || videoID <= 0 {
		return "", 0, entity.ErrInvalidVideo
	}
	return userID, videoID, nil
}

func (r *V1) userVideoMutationError(ctx *fiber.Ctx, err error, operation string) error {
	switch {
	case errors.Is(err, entity.ErrInvalidVideo):
		return errorResponse(ctx, http.StatusBadRequest, "invalid video")
	case errors.Is(err, entity.ErrVideoNotFound):
		return errorResponse(ctx, http.StatusNotFound, "video not found")
	default:
		r.l.Error(err, "restapi - v1 - user videos - "+operation)
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}

func userVideoRequest(ctx *fiber.Ctx) (string, int, int, error) {
	userID, ok := ctx.Locals("userID").(string)
	if !ok || userID == "" {
		return "", 0, 0, entity.ErrUserNotFound
	}
	page, limit := 1, 20
	var err error
	if raw := ctx.Query("page"); raw != "" {
		page, err = strconv.Atoi(raw)
		if err != nil || page <= 0 {
			return "", 0, 0, entity.ErrInvalidVideo
		}
	}
	if raw := ctx.Query("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit <= 0 || limit > 100 {
			return "", 0, 0, entity.ErrInvalidVideo
		}
	}
	return userID, page, limit, nil
}

func (r *V1) userVideoError(ctx *fiber.Ctx, err error, operation string) error {
	if errors.Is(err, entity.ErrInvalidVideo) {
		return errorResponse(ctx, http.StatusBadRequest, "invalid pagination")
	}
	r.l.Error(err, "restapi - v1 - user videos - "+operation)
	return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
}

func compactUserVideo(video entity.UserVideo) response.FeedVideo {
	return response.FeedVideo{ID: video.ID, Title: video.Title, YouTubeID: video.YouTubeID,
		ThumbnailURL: video.ThumbnailURL, DurationSeconds: video.DurationSeconds, LevelCode: video.LevelCode}
}

func buildWatchHistory(list entity.UserVideoList, page, limit int) response.WatchHistory {
	result := response.WatchHistory{Videos: make([]response.HistoryVideo, 0, len(list.Items)),
		Page: page, Limit: limit, Total: list.Total, TotalPages: totalPages(list.Total, limit)}
	for _, video := range list.Items {
		result.Videos = append(result.Videos, response.HistoryVideo{
			FeedVideo: compactUserVideo(video), LastPositionSeconds: video.LastPositionSeconds,
			LastWatchedAt: video.LastWatchedAt,
		})
	}
	return result
}

func buildWatchStatus(video entity.UserVideo, watched bool) response.WatchStatus {
	result := response.WatchStatus{Watched: watched}
	if watched {
		history := response.HistoryVideo{FeedVideo: compactUserVideo(video), LastPositionSeconds: video.LastPositionSeconds, LastWatchedAt: video.LastWatchedAt}
		result.Video = &history
	}
	return result
}

func buildWatchLater(list entity.UserVideoList, page, limit int) response.WatchLater {
	result := response.WatchLater{Videos: make([]response.SavedVideo, 0, len(list.Items)),
		Page: page, Limit: limit, Total: list.Total, TotalPages: totalPages(list.Total, limit)}
	for _, video := range list.Items {
		result.Videos = append(result.Videos, response.SavedVideo{FeedVideo: compactUserVideo(video), CreatedAt: video.SavedAt})
	}
	return result
}

func totalPages(total, limit int) int {
	if total == 0 {
		return 0
	}
	return (total + limit - 1) / limit
}
