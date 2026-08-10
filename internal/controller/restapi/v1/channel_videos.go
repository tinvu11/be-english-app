package v1

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary Get videos by channel
// @Description Return published videos of an active channel for the user's target language
// @ID user-channel-videos
// @Tags channels
// @Produce json
// @Param channelId path int true "Channel ID"
// @Param levelId query int false "Level ID"
// @Param search query string false "Search by video title"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Videos per page (1-100)" default(20)
// @Success 200 {object} response.ChannelVideos
// @Failure 400,401,404,500 {object} response.Error
// @Security BearerAuth
// @Router /channels/{channelId}/videos [get]
func (r *V1) channelVideos(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	channelID, page, limit, levelID, search, err := parseChannelVideosQuery(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video filter")
	}
	user, err := r.u.GetUser(ctx.UserContext(), userID)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "user not found")
		}
		r.l.Error(err, "restapi - v1 - channel videos - get user")
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
	if user.TargetLanguageID == nil {
		return errorResponse(ctx, http.StatusBadRequest, "user target language is not configured")
	}
	result, err := r.videos.ListVideos(ctx.UserContext(), entity.VideoFilter{
		LanguageID: user.TargetLanguageID, LevelID: levelID, ChannelID: &channelID,
		Status: entity.VideoStatusPublished, Search: search, OnlyActiveChannels: true,
		Limit: limit, Offset: (page - 1) * limit,
	})
	if err != nil {
		if errors.Is(err, entity.ErrInvalidVideo) {
			return errorResponse(ctx, http.StatusBadRequest, "invalid video filter")
		}
		r.l.Error(err, "restapi - v1 - channel videos - list videos")
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
	return ctx.Status(http.StatusOK).JSON(buildChannelVideos(result, page, limit))
}

func parseChannelVideosQuery(ctx *fiber.Ctx) (int, int, int, *int, string, error) {
	channelID, err := strconv.Atoi(ctx.Params("channelId"))
	if err != nil || channelID <= 0 {
		return 0, 0, 0, nil, "", entity.ErrInvalidVideo
	}
	page, limit := 1, 20
	if raw := ctx.Query("page"); raw != "" {
		page, err = strconv.Atoi(raw)
		if err != nil || page <= 0 {
			return 0, 0, 0, nil, "", entity.ErrInvalidVideo
		}
	}
	if raw := ctx.Query("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit <= 0 || limit > 100 {
			return 0, 0, 0, nil, "", entity.ErrInvalidVideo
		}
	}
	levelID, err := optionalPositiveInt(ctx.Query("levelId"))
	if err != nil {
		return 0, 0, 0, nil, "", err
	}
	search := strings.TrimSpace(ctx.Query("search"))
	if len(search) > 255 {
		return 0, 0, 0, nil, "", entity.ErrInvalidVideo
	}
	return channelID, page, limit, levelID, search, nil
}

func buildChannelVideos(list entity.VideoList, page, limit int) response.ChannelVideos {
	result := response.ChannelVideos{
		Videos: make([]response.FeedVideo, 0, len(list.Items)), Page: page, Limit: limit, Total: list.Total,
	}
	if list.Total > 0 {
		result.TotalPages = (list.Total + limit - 1) / limit
	}
	for _, video := range list.Items {
		result.Videos = append(result.Videos, response.FeedVideo{
			ID: video.ID, Title: video.Title, YouTubeID: video.YouTubeID, ThumbnailURL: video.ThumbnailURL,
			DurationSeconds: video.DurationSeconds, LevelCode: video.Level.Code,
		})
	}
	return result
}
