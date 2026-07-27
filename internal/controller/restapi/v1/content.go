package v1

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/evrone/go-clean-template/internal/controller/restapi/middleware"
	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/request"
	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

func registerContentRoutes(router fiber.Router, r *V1) {
	topics := router.Group("/topics")
	topics.Get("/", r.listTopics)
	topics.Get("/:id", r.getTopic)

	adminTopics := topics.Group("", middleware.AdminOnly())
	adminTopics.Post("/", r.createTopic)
	adminTopics.Put("/:id", r.updateTopic)
	adminTopics.Delete("/:id", r.deleteTopic)

	levels := router.Group("/levels")
	levels.Get("/", r.listLevels)
	levels.Get("/:id", r.getLevel)

	adminLevels := levels.Group("", middleware.AdminOnly())
	adminLevels.Post("/", r.createLevel)
	adminLevels.Put("/:id", r.updateLevel)
	adminLevels.Delete("/:id", r.deleteLevel)

	channels := router.Group("/channels")
	channels.Get("/", r.listChannels)
	channels.Get("/:id", r.getChannel)

	adminChannels := channels.Group("", middleware.AdminOnly())
	adminChannels.Post("/", r.createChannel)
	adminChannels.Put("/:id", r.updateChannel)
	adminChannels.Delete("/:id", r.deleteChannel)

	videos := router.Group("/videos")
	videos.Get("/", r.listVideos)
	videos.Get("/:id", r.getVideo)

	adminVideos := videos.Group("", middleware.AdminOnly())
	adminVideos.Post("/", r.createVideo)
	adminVideos.Put("/:id", r.updateVideo)
	adminVideos.Delete("/:id", r.deleteVideo)
}

func active(value *bool) bool {
	return value == nil || *value
}

func parseID(ctx *fiber.Ctx) (int64, error) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, errorResponse(ctx, http.StatusBadRequest, "invalid id")
	}

	return id, nil
}

func pagination(ctx *fiber.Ctx) (int, int) {
	limit, err := strconv.Atoi(ctx.Query("limit", "20"))
	if err != nil {
		limit = 20
	}
	offset, err := strconv.Atoi(ctx.Query("offset", "0"))
	if err != nil {
		offset = 0
	}

	return limit, offset
}

func (r *V1) bind(ctx *fiber.Ctx, body any) error {
	if err := ctx.BodyParser(body); err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}
	if err := r.v.Struct(body); err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	return nil
}

func (r *V1) contentError(ctx *fiber.Ctx, err error) error {
	r.l.Error(err, "restapi - v1 - content")
	switch {
	case errors.Is(err, entity.ErrContentNotFound):
		return errorResponse(ctx, http.StatusNotFound, "resource not found")
	case errors.Is(err, entity.ErrContentConflict):
		return errorResponse(ctx, http.StatusConflict, "resource already exists")
	case errors.Is(err, entity.ErrContentReferenced):
		return errorResponse(ctx, http.StatusConflict, "resource is referenced")
	case errors.Is(err, entity.ErrInvalidReference):
		return errorResponse(ctx, http.StatusBadRequest, "invalid reference")
	default:
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}

// @Summary Create topic
// @Tags topics
// @Accept json
// @Produce json
// @Param request body request.TopicRequest true "Topic"
// @Success 201 {object} entity.Topic
// @Failure 400,401,409,500 {object} response.Error
// @Security BearerAuth
// @Router /topics [post]
func (r *V1) createTopic(ctx *fiber.Ctx) error {
	var body request.TopicRequest
	if err := r.bind(ctx, &body); err != nil {
		return err
	}
	item, err := r.c.CreateTopic(ctx.UserContext(), entity.Topic{
		Name: body.Name, Slug: body.Slug, Icon: body.Icon, Description: body.Description,
		SortOrder: body.SortOrder, IsActive: active(body.IsActive),
	})
	if err != nil {
		return r.contentError(ctx, err)
	}

	return ctx.Status(http.StatusCreated).JSON(item)
}

// @Summary List topics
// @Tags topics
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} response.ContentList[entity.Topic]
// @Security BearerAuth
// @Router /topics [get]
func (r *V1) listTopics(ctx *fiber.Ctx) error {
	limit, offset := pagination(ctx)
	items, total, err := r.c.ListTopics(ctx.UserContext(), limit, offset)
	if err != nil {
		return r.contentError(ctx, err)
	}

	return ctx.JSON(response.ContentList[entity.Topic]{Items: items, Total: total})
}

// @Summary Get topic
// @Tags topics
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} entity.Topic
// @Failure 400,404,500 {object} response.Error
// @Security BearerAuth
// @Router /topics/{id} [get]
func (r *V1) getTopic(ctx *fiber.Ctx) error {
	id, err := parseID(ctx)
	if err != nil {
		return err
	}
	item, err := r.c.GetTopic(ctx.UserContext(), id)
	if err != nil {
		return r.contentError(ctx, err)
	}

	return ctx.JSON(item)
}

// @Summary Update topic
// @Tags topics
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param request body request.TopicRequest true "Topic"
// @Success 200 {object} entity.Topic
// @Failure 400,404,409,500 {object} response.Error
// @Security BearerAuth
// @Router /topics/{id} [put]
func (r *V1) updateTopic(ctx *fiber.Ctx) error {
	id, err := parseID(ctx)
	if err != nil {
		return err
	}
	var body request.TopicRequest
	if err = r.bind(ctx, &body); err != nil {
		return err
	}
	item, err := r.c.UpdateTopic(ctx.UserContext(), id, entity.Topic{
		Name: body.Name, Slug: body.Slug, Icon: body.Icon, Description: body.Description,
		SortOrder: body.SortOrder, IsActive: active(body.IsActive),
	})
	if err != nil {
		return r.contentError(ctx, err)
	}

	return ctx.JSON(item)
}

// @Summary Delete topic
// @Tags topics
// @Param id path int true "ID"
// @Success 204
// @Failure 400,404,409,500 {object} response.Error
// @Security BearerAuth
// @Router /topics/{id} [delete]
func (r *V1) deleteTopic(ctx *fiber.Ctx) error {
	id, err := parseID(ctx)
	if err != nil {
		return err
	}
	if err = r.c.DeleteTopic(ctx.UserContext(), id); err != nil {
		return r.contentError(ctx, err)
	}

	return ctx.SendStatus(http.StatusNoContent)
}

// @Summary Create level
// @Tags levels
// @Accept json
// @Produce json
// @Param request body request.LevelRequest true "Level"
// @Success 201 {object} entity.Level
// @Failure 400,409,500 {object} response.Error
// @Security BearerAuth
// @Router /levels [post]
func (r *V1) createLevel(ctx *fiber.Ctx) error {
	var body request.LevelRequest
	if err := r.bind(ctx, &body); err != nil {
		return err
	}
	item, err := r.c.CreateLevel(ctx.UserContext(), entity.Level{Code: body.Code, Name: body.Name, SortOrder: body.SortOrder})
	if err != nil {
		return r.contentError(ctx, err)
	}

	return ctx.Status(http.StatusCreated).JSON(item)
}

// @Summary List levels
// @Tags levels
// @Produce json
// @Success 200 {object} response.ContentList[entity.Level]
// @Security BearerAuth
// @Router /levels [get]
func (r *V1) listLevels(ctx *fiber.Ctx) error {
	limit, offset := pagination(ctx)
	items, total, err := r.c.ListLevels(ctx.UserContext(), limit, offset)
	if err != nil {
		return r.contentError(ctx, err)
	}

	return ctx.JSON(response.ContentList[entity.Level]{Items: items, Total: total})
}

// @Summary Get level
// @Tags levels
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} entity.Level
// @Failure 400,404,500 {object} response.Error
// @Security BearerAuth
// @Router /levels/{id} [get]
func (r *V1) getLevel(ctx *fiber.Ctx) error {
	id, err := parseID(ctx)
	if err != nil {
		return err
	}
	item, err := r.c.GetLevel(ctx.UserContext(), id)
	if err != nil {
		return r.contentError(ctx, err)
	}

	return ctx.JSON(item)
}

// @Summary Update level
// @Tags levels
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param request body request.LevelRequest true "Level"
// @Success 200 {object} entity.Level
// @Failure 400,404,409,500 {object} response.Error
// @Security BearerAuth
// @Router /levels/{id} [put]
func (r *V1) updateLevel(ctx *fiber.Ctx) error {
	id, err := parseID(ctx)
	if err != nil {
		return err
	}
	var body request.LevelRequest
	if err = r.bind(ctx, &body); err != nil {
		return err
	}
	item, err := r.c.UpdateLevel(ctx.UserContext(), id, entity.Level{Code: body.Code, Name: body.Name, SortOrder: body.SortOrder})
	if err != nil {
		return r.contentError(ctx, err)
	}

	return ctx.JSON(item)
}

// @Summary Delete level
// @Tags levels
// @Param id path int true "ID"
// @Success 204
// @Failure 400,404,409,500 {object} response.Error
// @Security BearerAuth
// @Router /levels/{id} [delete]
func (r *V1) deleteLevel(ctx *fiber.Ctx) error {
	id, err := parseID(ctx)
	if err != nil {
		return err
	}
	if err = r.c.DeleteLevel(ctx.UserContext(), id); err != nil {
		return r.contentError(ctx, err)
	}

	return ctx.SendStatus(http.StatusNoContent)
}

// @Summary Create channel
// @Tags channels
// @Accept json
// @Produce json
// @Param request body request.ChannelRequest true "Channel"
// @Success 201 {object} entity.Channel
// @Failure 400,409,500 {object} response.Error
// @Security BearerAuth
// @Router /channels [post]
func (r *V1) createChannel(ctx *fiber.Ctx) error {
	var body request.ChannelRequest
	if err := r.bind(ctx, &body); err != nil {
		return err
	}
	item, err := r.c.CreateChannel(ctx.UserContext(), entity.Channel{
		ChannelID: body.ChannelID, ChannelName: body.ChannelName, ThumbnailURL: body.ThumbnailURL,
	})
	if err != nil {
		return r.contentError(ctx, err)
	}

	return ctx.Status(http.StatusCreated).JSON(item)
}

// @Summary List channels
// @Tags channels
// @Produce json
// @Success 200 {object} response.ContentList[entity.Channel]
// @Security BearerAuth
// @Router /channels [get]
func (r *V1) listChannels(ctx *fiber.Ctx) error {
	limit, offset := pagination(ctx)
	items, total, err := r.c.ListChannels(ctx.UserContext(), limit, offset)
	if err != nil {
		return r.contentError(ctx, err)
	}

	return ctx.JSON(response.ContentList[entity.Channel]{Items: items, Total: total})
}

// @Summary Get channel
// @Tags channels
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} entity.Channel
// @Failure 400,404,500 {object} response.Error
// @Security BearerAuth
// @Router /channels/{id} [get]
func (r *V1) getChannel(ctx *fiber.Ctx) error {
	id, err := parseID(ctx)
	if err != nil {
		return err
	}
	item, err := r.c.GetChannel(ctx.UserContext(), id)
	if err != nil {
		return r.contentError(ctx, err)
	}

	return ctx.JSON(item)
}

// @Summary Update channel
// @Tags channels
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param request body request.ChannelRequest true "Channel"
// @Success 200 {object} entity.Channel
// @Failure 400,404,409,500 {object} response.Error
// @Security BearerAuth
// @Router /channels/{id} [put]
func (r *V1) updateChannel(ctx *fiber.Ctx) error {
	id, err := parseID(ctx)
	if err != nil {
		return err
	}
	var body request.ChannelRequest
	if err = r.bind(ctx, &body); err != nil {
		return err
	}
	item, err := r.c.UpdateChannel(ctx.UserContext(), id, entity.Channel{
		ChannelID: body.ChannelID, ChannelName: body.ChannelName, ThumbnailURL: body.ThumbnailURL,
	})
	if err != nil {
		return r.contentError(ctx, err)
	}

	return ctx.JSON(item)
}

// @Summary Delete channel
// @Tags channels
// @Param id path int true "ID"
// @Success 204
// @Failure 400,404,409,500 {object} response.Error
// @Security BearerAuth
// @Router /channels/{id} [delete]
func (r *V1) deleteChannel(ctx *fiber.Ctx) error {
	id, err := parseID(ctx)
	if err != nil {
		return err
	}
	if err = r.c.DeleteChannel(ctx.UserContext(), id); err != nil {
		return r.contentError(ctx, err)
	}

	return ctx.SendStatus(http.StatusNoContent)
}

func video(body request.VideoRequest) entity.Video {
	return entity.Video{
		VideoID: body.VideoID, ChannelID: body.ChannelID, Title: body.Title, Description: body.Description,
		ThumbnailURL: body.ThumbnailURL, ViewCount: body.ViewCount, Duration: body.Duration,
		SortOrder: body.SortOrder, IsActive: active(body.IsActive), PublishedAt: body.PublishedAt,
		TopicIDs: body.TopicIDs, LevelIDs: body.LevelIDs,
	}
}

// @Summary Create video
// @Tags videos
// @Accept json
// @Produce json
// @Param request body request.VideoRequest true "Video"
// @Success 201 {object} entity.Video
// @Failure 400,409,500 {object} response.Error
// @Security BearerAuth
// @Router /videos [post]
func (r *V1) createVideo(ctx *fiber.Ctx) error {
	var body request.VideoRequest
	if err := r.bind(ctx, &body); err != nil {
		return err
	}
	item, err := r.c.CreateVideo(ctx.UserContext(), video(body))
	if err != nil {
		return r.contentError(ctx, err)
	}

	return ctx.Status(http.StatusCreated).JSON(item)
}

// @Summary List videos
// @Tags videos
// @Produce json
// @Success 200 {object} response.ContentList[entity.Video]
// @Security BearerAuth
// @Router /videos [get]
func (r *V1) listVideos(ctx *fiber.Ctx) error {
	limit, offset := pagination(ctx)
	items, total, err := r.c.ListVideos(ctx.UserContext(), limit, offset)
	if err != nil {
		return r.contentError(ctx, err)
	}

	return ctx.JSON(response.ContentList[entity.Video]{Items: items, Total: total})
}

// @Summary Get video
// @Tags videos
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} entity.Video
// @Failure 400,404,500 {object} response.Error
// @Security BearerAuth
// @Router /videos/{id} [get]
func (r *V1) getVideo(ctx *fiber.Ctx) error {
	id, err := parseID(ctx)
	if err != nil {
		return err
	}
	item, err := r.c.GetVideo(ctx.UserContext(), id)
	if err != nil {
		return r.contentError(ctx, err)
	}

	return ctx.JSON(item)
}

// @Summary Update video
// @Tags videos
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param request body request.VideoRequest true "Video"
// @Success 200 {object} entity.Video
// @Failure 400,404,409,500 {object} response.Error
// @Security BearerAuth
// @Router /videos/{id} [put]
func (r *V1) updateVideo(ctx *fiber.Ctx) error {
	id, err := parseID(ctx)
	if err != nil {
		return err
	}
	var body request.VideoRequest
	if err = r.bind(ctx, &body); err != nil {
		return err
	}
	item, err := r.c.UpdateVideo(ctx.UserContext(), id, video(body))
	if err != nil {
		return r.contentError(ctx, err)
	}

	return ctx.JSON(item)
}

// @Summary Delete video
// @Tags videos
// @Param id path int true "ID"
// @Success 204
// @Failure 400,404,500 {object} response.Error
// @Security BearerAuth
// @Router /videos/{id} [delete]
func (r *V1) deleteVideo(ctx *fiber.Ctx) error {
	id, err := parseID(ctx)
	if err != nil {
		return err
	}
	if err = r.c.DeleteVideo(ctx.UserContext(), id); err != nil {
		return r.contentError(ctx, err)
	}

	return ctx.SendStatus(http.StatusNoContent)
}
