package admin

import (
	"errors"
	"net/http"
	"strconv"

	adminrequest "github.com/evrone/go-clean-template/internal/controller/restapi/v1/admin/request"
	_ "github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary     List levels
// @Description Return levels, optionally filtered by their source language
// @ID          admin-list-levels
// @Tags        admin-levels
// @Produce     json
// @Param       language_id query int false "Source language ID"
// @Success     200 {array} entity.Level
// @Failure     400 {object} errorDoc
// @Failure     401 {object} errorDoc
// @Failure     403 {object} errorDoc
// @Failure     500 {object} errorDoc
// @Security    BearerAuth
// @Router      /admin/levels [get]
func (ctrl *controller) listLevels(ctx *fiber.Ctx) error {
	var languageID *int
	if raw := ctx.Query("language_id"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			return errorResponse(ctx, http.StatusBadRequest, "invalid language_id")
		}
		languageID = &parsed
	}

	levels, err := ctrl.levels.ListLevels(ctx.UserContext(), languageID)
	if err != nil {
		return ctrl.levelError(ctx, err)
	}

	return ctx.Status(http.StatusOK).JSON(levels)
}

// @Summary     Create level
// @Description Create a level; multilingual translations are optional
// @ID          admin-create-level
// @Tags        admin-levels
// @Accept      json
// @Produce     json
// @Param       request body adminrequest.SaveLevel true "Level and translations"
// @Success     201 {object} entity.Level
// @Failure     400 {object} errorDoc
// @Failure     401 {object} errorDoc
// @Failure     403 {object} errorDoc
// @Failure     409 {object} errorDoc
// @Failure     500 {object} errorDoc
// @Security    BearerAuth
// @Router      /admin/levels [post]
func (ctrl *controller) createLevel(ctx *fiber.Ctx) error {
	body, err := ctrl.parseLevel(ctx)
	if err != nil {
		return err
	}

	level, err := ctrl.levels.CreateLevel(ctx.UserContext(), levelFromRequest(body))
	if err != nil {
		return ctrl.levelError(ctx, err)
	}

	return ctx.Status(http.StatusCreated).JSON(level)
}

// @Summary     Update level
// @Description Replace a level and its optional multilingual translations
// @ID          admin-update-level
// @Tags        admin-levels
// @Accept      json
// @Produce     json
// @Param       id path int true "Level ID"
// @Param       request body adminrequest.SaveLevel true "Level and translations"
// @Success     200 {object} entity.Level
// @Failure     400 {object} errorDoc
// @Failure     401 {object} errorDoc
// @Failure     403 {object} errorDoc
// @Failure     404 {object} errorDoc
// @Failure     409 {object} errorDoc
// @Failure     500 {object} errorDoc
// @Security    BearerAuth
// @Router      /admin/levels/{id} [put]
func (ctrl *controller) updateLevel(ctx *fiber.Ctx) error {
	id, err := positiveID(ctx.Params("id"))
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid level id")
	}
	body, err := ctrl.parseLevel(ctx)
	if err != nil {
		return err
	}

	level, err := ctrl.levels.UpdateLevel(ctx.UserContext(), id, levelFromRequest(body))
	if err != nil {
		return ctrl.levelError(ctx, err)
	}

	return ctx.Status(http.StatusOK).JSON(level)
}

// @Summary     Delete level
// @Description Delete a level only when no video references it
// @ID          admin-delete-level
// @Tags        admin-levels
// @Param       id path int true "Level ID"
// @Success     204
// @Failure     400 {object} errorDoc
// @Failure     401 {object} errorDoc
// @Failure     403 {object} errorDoc
// @Failure     404 {object} errorDoc
// @Failure     409 {object} errorDoc
// @Failure     500 {object} errorDoc
// @Security    BearerAuth
// @Router      /admin/levels/{id} [delete]
func (ctrl *controller) deleteLevel(ctx *fiber.Ctx) error {
	id, err := positiveID(ctx.Params("id"))
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid level id")
	}
	if err = ctrl.levels.DeleteLevel(ctx.UserContext(), id); err != nil {
		return ctrl.levelError(ctx, err)
	}

	return ctx.SendStatus(http.StatusNoContent)
}

func (ctrl *controller) parseLevel(ctx *fiber.Ctx) (adminrequest.SaveLevel, error) {
	var body adminrequest.SaveLevel
	if err := ctx.BodyParser(&body); err != nil {
		return body, errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}
	if err := ctrl.validate.Struct(body); err != nil {
		return body, errorResponse(ctx, http.StatusBadRequest, "invalid level")
	}

	return body, nil
}

func levelFromRequest(body adminrequest.SaveLevel) entity.Level {
	translations := make([]entity.LevelTranslation, len(body.Translations))
	for index, translation := range body.Translations {
		translations[index] = entity.LevelTranslation{LanguageID: translation.LanguageID, Name: translation.Name}
	}

	return entity.Level{Code: body.Code, LanguageID: body.LanguageID, Translations: translations}
}

func positiveID(raw string) (int, error) {
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		return 0, entity.ErrInvalidLevel
	}

	return id, nil
}

func (ctrl *controller) levelError(ctx *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, entity.ErrInvalidLevel), errors.Is(err, entity.ErrInvalidReference):
		return errorResponse(ctx, http.StatusBadRequest, "invalid level or language reference")
	case errors.Is(err, entity.ErrLevelNotFound):
		return errorResponse(ctx, http.StatusNotFound, "level not found")
	case errors.Is(err, entity.ErrLevelExists):
		return errorResponse(ctx, http.StatusConflict, "level code already exists for language")
	case errors.Is(err, entity.ErrLevelReferenced):
		return errorResponse(ctx, http.StatusConflict, "level is referenced by videos")
	default:
		ctrl.log.Error(err, "restapi - v1 - admin - level")

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}
