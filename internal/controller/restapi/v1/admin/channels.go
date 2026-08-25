package admin

import (
	"errors"
	"net/http"

	adminrequest "github.com/evrone/go-clean-template/internal/controller/restapi/v1/admin/request"
	_ "github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary List YouTube channels
// @Tags admin-channels
// @Produce json
// @Success 200 {array} entity.Channel
// @Failure 401,403,500 {object} errorDoc
// @Security BearerAuth
// @Router /admin/channels [get]
func (ctrl *controller) listChannels(ctx *fiber.Ctx) error {
	items, err := ctrl.channels.ListChannels(ctx.UserContext())
	if err != nil {
		return ctrl.channelError(ctx, err)
	}
	return ctx.JSON(items)
}

// @Summary Create YouTube channel
// @Tags admin-channels
// @Accept json
// @Produce json
// @Param request body adminrequest.SaveChannel true "Channel"
// @Success 201 {object} entity.Channel
// @Failure 400,401,403,409,500 {object} errorDoc
// @Security BearerAuth
// @Router /admin/channels [post]
func (ctrl *controller) createChannel(ctx *fiber.Ctx) error {
	body, err := ctrl.parseChannel(ctx)
	if err != nil {
		return err
	}
	item, err := ctrl.channels.CreateChannel(ctx.UserContext(), channelFromRequest(body))
	if err != nil {
		return ctrl.channelError(ctx, err)
	}
	return ctx.Status(http.StatusCreated).JSON(item)
}

// @Summary Update YouTube channel
// @Tags admin-channels
// @Accept json
// @Produce json
// @Param id path int true "Channel ID"
// @Param request body adminrequest.SaveChannel true "Channel"
// @Success 200 {object} entity.Channel
// @Failure 400,401,403,404,409,500 {object} errorDoc
// @Security BearerAuth
// @Router /admin/channels/{id} [put]
func (ctrl *controller) updateChannel(ctx *fiber.Ctx) error {
	id, err := positiveID(ctx.Params("id"))
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid channel id")
	}
	body, err := ctrl.parseChannel(ctx)
	if err != nil {
		return err
	}
	item, err := ctrl.channels.UpdateChannel(ctx.UserContext(), id, channelFromRequest(body))
	if err != nil {
		return ctrl.channelError(ctx, err)
	}
	return ctx.JSON(item)
}

// @Summary Delete YouTube channel
// @Description Delete a channel only when no video references it
// @Tags admin-channels
// @Param id path int true "Channel ID"
// @Success 204
// @Failure 400,401,403,404,409,500 {object} errorDoc
// @Security BearerAuth
// @Router /admin/channels/{id} [delete]
func (ctrl *controller) deleteChannel(ctx *fiber.Ctx) error {
	id, err := positiveID(ctx.Params("id"))
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid channel id")
	}
	if err = ctrl.channels.DeleteChannel(ctx.UserContext(), id); err != nil {
		return ctrl.channelError(ctx, err)
	}
	return ctx.SendStatus(http.StatusNoContent)
}

func (ctrl *controller) parseChannel(ctx *fiber.Ctx) (adminrequest.SaveChannel, error) {
	var body adminrequest.SaveChannel
	if err := ctx.BodyParser(&body); err != nil {
		return body, errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}
	if err := ctrl.validate.Struct(body); err != nil {
		return body, errorResponse(ctx, http.StatusBadRequest, "invalid channel")
	}
	return body, nil
}

func channelFromRequest(body adminrequest.SaveChannel) entity.Channel {
	return entity.Channel{ChannelYouTubeID: body.ChannelYouTubeID, ChannelName: body.ChannelName,
		AvatarURL: body.AvatarURL, IsActive: *body.IsActive}
}

func (ctrl *controller) channelError(ctx *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, entity.ErrInvalidChannel):
		return errorResponse(ctx, http.StatusBadRequest, "invalid channel")
	case errors.Is(err, entity.ErrChannelNotFound):
		return errorResponse(ctx, http.StatusNotFound, "channel not found")
	case errors.Is(err, entity.ErrChannelExists):
		return errorResponse(ctx, http.StatusConflict, "YouTube channel already exists")
	case errors.Is(err, entity.ErrChannelReferenced):
		return errorResponse(ctx, http.StatusConflict, "channel is referenced by videos")
	default:
		ctrl.log.Error(err, "restapi - v1 - admin - channel")
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}
