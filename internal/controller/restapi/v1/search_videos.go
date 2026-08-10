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

// @Summary Search system videos
// @Description Search published system videos by title across active learnable languages
// @ID user-search-videos
// @Tags videos
// @Produce json
// @Param search query string true "Video title keyword"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Videos per page (1-100)" default(20)
// @Success 200 {object} response.SearchVideos
// @Failure 400,401,500 {object} response.Error
// @Security BearerAuth
// @Router /videos/search [get]
func (r *V1) searchVideos(ctx *fiber.Ctx) error {
	search, page, limit, err := parseVideoSearch(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video search")
	}
	result, err := r.videos.ListVideos(ctx.UserContext(), entity.VideoFilter{
		Status: entity.VideoStatusPublished, Search: search, OnlyActiveChannels: true,
		OnlyLearnableLanguages: true, Limit: limit, Offset: (page - 1) * limit,
	})
	if err != nil {
		if errors.Is(err, entity.ErrInvalidVideo) {
			return errorResponse(ctx, http.StatusBadRequest, "invalid video search")
		}
		r.l.Error(err, "restapi - v1 - search videos")
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
	return ctx.Status(http.StatusOK).JSON(buildSearchVideos(result, page, limit))
}

func parseVideoSearch(ctx *fiber.Ctx) (string, int, int, error) {
	search := strings.TrimSpace(ctx.Query("search"))
	if search == "" || len(search) > 255 {
		return "", 0, 0, entity.ErrInvalidVideo
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
	return search, page, limit, nil
}

func buildSearchVideos(list entity.VideoList, page, limit int) response.SearchVideos {
	result := response.SearchVideos{Videos: make([]response.SearchVideo, 0, len(list.Items)),
		Page: page, Limit: limit, Total: list.Total, TotalPages: totalPages(list.Total, limit)}
	for _, video := range list.Items {
		result.Videos = append(result.Videos, response.SearchVideo{
			FeedVideo: response.FeedVideo{ID: video.ID, Title: video.Title, YouTubeID: video.YouTubeID,
				ThumbnailURL: video.ThumbnailURL, DurationSeconds: video.DurationSeconds, LevelCode: video.Level.Code},
			Language: response.SearchVideoLanguage{ID: video.Language.ID, Code: video.Language.Code,
				Name: video.Language.Name, Flag: video.Language.FlagEmoji},
		})
	}
	return result
}
