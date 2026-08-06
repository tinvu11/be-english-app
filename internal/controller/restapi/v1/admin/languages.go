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

// @Summary     List languages
// @Description Return all platform languages ordered by code
// @ID          admin-list-languages
// @Tags        admin-languages
// @Produce     json
// @Success     200 {array} entity.Language
// @Failure     401 {object} response.Error
// @Failure     403 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /admin/languages [get]
func (ctrl *controller) listLanguages(ctx *fiber.Ctx) error {
	languages, err := ctrl.languages.ListLanguages(ctx.UserContext())
	if err != nil {
		ctrl.log.Error(err, "restapi - v1 - admin - list languages")

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(languages)
}

// @Summary     Create language
// @Description Add a new platform language; its code is immutable after creation
// @ID          admin-create-language
// @Tags        admin-languages
// @Accept      json
// @Produce     json
// @Param       request body adminrequest.CreateLanguage true "Language"
// @Success     201 {object} entity.Language
// @Failure     400 {object} response.Error
// @Failure     401 {object} response.Error
// @Failure     403 {object} response.Error
// @Failure     409 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /admin/languages [post]
func (ctrl *controller) createLanguage(ctx *fiber.Ctx) error {
	var body adminrequest.CreateLanguage
	if err := ctx.BodyParser(&body); err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}
	if err := ctrl.validate.Struct(body); err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid language")
	}

	language, err := ctrl.languages.CreateLanguage(ctx.UserContext(), body.Code, body.Name)
	if err != nil {
		return ctrl.languageError(ctx, err)
	}

	return ctx.Status(http.StatusCreated).JSON(language)
}

// @Summary     Update language
// @Description Update a language name and active status
// @ID          admin-update-language
// @Tags        admin-languages
// @Accept      json
// @Produce     json
// @Param       id path int true "Language ID"
// @Param       request body adminrequest.UpdateLanguage true "Language changes"
// @Success     200 {object} entity.Language
// @Failure     400 {object} response.Error
// @Failure     401 {object} response.Error
// @Failure     403 {object} response.Error
// @Failure     404 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /admin/languages/{id} [put]
func (ctrl *controller) updateLanguage(ctx *fiber.Ctx) error {
	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil || id <= 0 {
		return errorResponse(ctx, http.StatusBadRequest, "invalid language id")
	}

	var body adminrequest.UpdateLanguage
	if err = ctx.BodyParser(&body); err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}
	if err = ctrl.validate.Struct(body); err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid language")
	}

	language, err := ctrl.languages.UpdateLanguage(ctx.UserContext(), id, body.Name, *body.IsActive)
	if err != nil {
		return ctrl.languageError(ctx, err)
	}

	return ctx.Status(http.StatusOK).JSON(language)
}

func (ctrl *controller) languageError(ctx *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, entity.ErrInvalidLanguage):
		return errorResponse(ctx, http.StatusBadRequest, "invalid language")
	case errors.Is(err, entity.ErrLanguageNotFound):
		return errorResponse(ctx, http.StatusNotFound, "language not found")
	case errors.Is(err, entity.ErrLanguageExists):
		return errorResponse(ctx, http.StatusConflict, "language code already exists")
	default:
		ctrl.log.Error(err, "restapi - v1 - admin - language")

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}
