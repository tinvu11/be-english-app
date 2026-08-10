package v1

import (
	"errors"
	"net/http"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/request"
	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// userErrorDoc keeps the Swagger response type discoverable in this file.
type userErrorDoc = response.Error

// @Summary     Get profile
// @Description Get current user profile
// @ID          profile
// @Tags        user
// @Produce     json
// @Success     200 {object} response.UserProfile
// @Failure     401 {object} response.Error
// @Failure     404 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /user/profile [get]
func (r *V1) profile(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	user, err := r.u.GetUser(ctx.UserContext(), userID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - profile")

		if errors.Is(err, entity.ErrUserNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "user not found")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	profile, err := r.profileResponse(ctx, user)
	if err != nil {
		r.l.Error(err, "restapi - v1 - profile - list languages")
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
	return ctx.Status(http.StatusOK).JSON(profile)
}

// @Summary Update user languages
// @Description Update the native translation language and target learning language of the current user
// @ID user-update-languages
// @Tags user
// @Accept json
// @Produce json
// @Param request body request.UpdateUserLanguages true "Language choices"
// @Success 200 {object} response.UserProfile
// @Failure 400,401,404,500 {object} response.Error
// @Security BearerAuth
// @Router /user/languages [put]
func (r *V1) updateUserLanguages(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	var body request.UpdateUserLanguages
	if err := ctx.BodyParser(&body); err != nil || r.v.Struct(body) != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid user languages")
	}
	user, err := r.u.UpdateLanguages(ctx.UserContext(), userID, body.NativeLanguageID, body.TargetLanguageID)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrInvalidLanguage), errors.Is(err, entity.ErrInvalidReference):
			return errorResponse(ctx, http.StatusBadRequest, "invalid user languages")
		case errors.Is(err, entity.ErrLanguageNotFound):
			return errorResponse(ctx, http.StatusNotFound, "language not found")
		case errors.Is(err, entity.ErrUserNotFound):
			return errorResponse(ctx, http.StatusNotFound, "user not found")
		default:
			r.l.Error(err, "restapi - v1 - update user languages")
			return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
		}
	}
	profile, err := r.profileResponse(ctx, user)
	if err != nil {
		r.l.Error(err, "restapi - v1 - update user languages - list languages")
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
	return ctx.Status(http.StatusOK).JSON(profile)
}

func (r *V1) profileResponse(ctx *fiber.Ctx, user entity.User) (response.UserProfile, error) {
	result := response.UserProfile{ID: user.ID, Username: user.Username, Email: user.Email, AvatarURL: user.AvatarURL}
	languages, err := r.languages.ListLanguages(ctx.UserContext())
	if err != nil {
		return result, err
	}
	for _, language := range languages {
		item := response.OnboardingLanguage{ID: language.ID, Code: language.Code, Name: language.Name, Flag: language.FlagEmoji}
		if user.NativeLanguageID != nil && language.ID == *user.NativeLanguageID {
			copy := item
			result.NativeLanguage = &copy
		}
		if user.TargetLanguageID != nil && language.ID == *user.TargetLanguageID {
			copy := item
			result.TargetLanguage = &copy
		}
	}
	return result, nil
}
