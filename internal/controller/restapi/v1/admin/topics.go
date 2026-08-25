package admin

import (
	"errors"
	"net/http"

	adminrequest "github.com/evrone/go-clean-template/internal/controller/restapi/v1/admin/request"
	_ "github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary List topics
// @Tags admin-topics
// @Produce json
// @Success 200 {array} entity.Topic
// @Failure 401,403,500 {object} errorDoc
// @Security BearerAuth
// @Router /admin/topics [get]
func (ctrl *controller) listTopics(ctx *fiber.Ctx) error {
	items, err := ctrl.topics.ListTopics(ctx.UserContext())
	if err != nil {
		return ctrl.topicError(ctx, err)
	}
	return ctx.JSON(items)
}

// @Summary Create topic with translations
// @Tags admin-topics
// @Accept json
// @Produce json
// @Param request body adminrequest.SaveTopic true "Topic"
// @Success 201 {object} entity.Topic
// @Failure 400,401,403,409,500 {object} errorDoc
// @Security BearerAuth
// @Router /admin/topics [post]
func (ctrl *controller) createTopic(ctx *fiber.Ctx) error {
	body, err := ctrl.parseTopic(ctx)
	if err != nil {
		return err
	}
	item, err := ctrl.topics.CreateTopic(ctx.UserContext(), topicFromRequest(body))
	if err != nil {
		return ctrl.topicError(ctx, err)
	}
	return ctx.Status(http.StatusCreated).JSON(item)
}

// @Summary Update topic and translations
// @Tags admin-topics
// @Accept json
// @Produce json
// @Param id path int true "Topic ID"
// @Param request body adminrequest.SaveTopic true "Topic"
// @Success 200 {object} entity.Topic
// @Failure 400,401,403,404,409,500 {object} errorDoc
// @Security BearerAuth
// @Router /admin/topics/{id} [put]
func (ctrl *controller) updateTopic(ctx *fiber.Ctx) error {
	id, err := positiveID(ctx.Params("id"))
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid topic id")
	}
	body, err := ctrl.parseTopic(ctx)
	if err != nil {
		return err
	}
	item, err := ctrl.topics.UpdateTopic(ctx.UserContext(), id, topicFromRequest(body))
	if err != nil {
		return ctrl.topicError(ctx, err)
	}
	return ctx.JSON(item)
}

// @Summary Enable or disable topic visibility
// @Tags admin-topics
// @Accept json
// @Produce json
// @Param id path int true "Topic ID"
// @Param request body adminrequest.TopicStatus true "Visibility"
// @Success 200 {object} entity.Topic
// @Failure 400,401,403,404,500 {object} errorDoc
// @Security BearerAuth
// @Router /admin/topics/{id}/status [patch]
func (ctrl *controller) setTopicActive(ctx *fiber.Ctx) error {
	id, err := positiveID(ctx.Params("id"))
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid topic id")
	}
	var body adminrequest.TopicStatus
	if err = ctx.BodyParser(&body); err != nil || ctrl.validate.Struct(body) != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid topic status")
	}
	item, err := ctrl.topics.SetTopicActive(ctx.UserContext(), id, *body.IsActive)
	if err != nil {
		return ctrl.topicError(ctx, err)
	}
	return ctx.JSON(item)
}

// @Summary Delete topic
// @Description Delete a topic and its translations only when no video references it
// @Tags admin-topics
// @Param id path int true "Topic ID"
// @Success 204
// @Failure 400,401,403,404,409,500 {object} errorDoc
// @Security BearerAuth
// @Router /admin/topics/{id} [delete]
func (ctrl *controller) deleteTopic(ctx *fiber.Ctx) error {
	id, err := positiveID(ctx.Params("id"))
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid topic id")
	}
	if err = ctrl.topics.DeleteTopic(ctx.UserContext(), id); err != nil {
		return ctrl.topicError(ctx, err)
	}
	return ctx.SendStatus(http.StatusNoContent)
}

func (ctrl *controller) parseTopic(ctx *fiber.Ctx) (adminrequest.SaveTopic, error) {
	var body adminrequest.SaveTopic
	if err := ctx.BodyParser(&body); err != nil {
		return body, errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}
	if err := ctrl.validate.Struct(body); err != nil {
		return body, errorResponse(ctx, http.StatusBadRequest, "invalid topic")
	}
	return body, nil
}

func topicFromRequest(body adminrequest.SaveTopic) entity.Topic {
	translations := make([]entity.TopicTranslation, len(body.Translations))
	for i, v := range body.Translations {
		translations[i] = entity.TopicTranslation{LanguageID: v.LanguageID, Name: v.Name}
	}
	return entity.Topic{Slug: body.Slug, IconURL: body.IconURL, IsActive: *body.IsActive, Translations: translations}
}

func (ctrl *controller) topicError(ctx *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, entity.ErrInvalidTopic), errors.Is(err, entity.ErrInvalidReference):
		return errorResponse(ctx, http.StatusBadRequest, "invalid topic or language reference")
	case errors.Is(err, entity.ErrTopicNotFound):
		return errorResponse(ctx, http.StatusNotFound, "topic not found")
	case errors.Is(err, entity.ErrTopicExists):
		return errorResponse(ctx, http.StatusConflict, "topic slug already exists")
	case errors.Is(err, entity.ErrTopicReferenced):
		return errorResponse(ctx, http.StatusConflict, "topic is referenced by videos")
	default:
		ctrl.log.Error(err, "restapi - v1 - admin - topic")
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}
