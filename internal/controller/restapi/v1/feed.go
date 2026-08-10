package v1

import (
	"errors"
	"net/http"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary Get home feed
// @Description Return published videos for the user's target language, grouped by channel
// @ID user-home-feed
// @Tags home
// @Produce json
// @Param topicId query int false "Topic ID"
// @Param levelId query int false "Level ID"
// @Success 200 {object} response.HomeFeed
// @Failure 400,401,404,500 {object} response.Error
// @Security BearerAuth
// @Router /home/feed [get]
func (r *V1) homeFeed(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	topicID, err := optionalPositiveInt(ctx.Query("topicId"))
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid feed filter")
	}
	levelID, err := optionalPositiveInt(ctx.Query("levelId"))
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid feed filter")
	}
	user, err := r.u.GetUser(ctx.UserContext(), userID)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "user not found")
		}
		r.l.Error(err, "restapi - v1 - home feed - get user")
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
	if user.TargetLanguageID == nil {
		return errorResponse(ctx, http.StatusBadRequest, "user target language is not configured")
	}
	result, err := r.videos.ListVideos(ctx.UserContext(), entity.VideoFilter{
		LanguageID: user.TargetLanguageID, LevelID: levelID, TopicID: topicID,
		Status: entity.VideoStatusPublished, OnlyActiveChannels: true,
		PerChannelLimit: 10, Limit: 10,
	})
	if err != nil {
		if errors.Is(err, entity.ErrInvalidVideo) {
			return errorResponse(ctx, http.StatusBadRequest, "invalid feed filter")
		}
		r.l.Error(err, "restapi - v1 - home feed - list videos")
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
	return ctx.Status(http.StatusOK).JSON(buildHomeFeed(result))
}

func buildHomeFeed(list entity.VideoList) response.HomeFeed {
	feed := response.HomeFeed{Sections: make([]response.FeedSection, 0)}
	positions := make(map[int]int)
	for _, video := range list.Items {
		position, exists := positions[video.Channel.ID]
		if !exists {
			position = len(feed.Sections)
			positions[video.Channel.ID] = position
			feed.Sections = append(feed.Sections, response.FeedSection{
				Channel: response.FeedChannel{ID: video.Channel.ID, Name: video.Channel.Name, AvatarURL: video.Channel.AvatarURL},
				Videos:  make([]response.FeedVideo, 0),
			})
		}
		feed.Sections[position].Videos = append(feed.Sections[position].Videos, response.FeedVideo{
			ID: video.ID, Title: video.Title, YouTubeID: video.YouTubeID, ThumbnailURL: video.ThumbnailURL,
			DurationSeconds: video.DurationSeconds, LevelCode: video.Level.Code,
		})
	}
	return feed
}
