package v1

import (
	"errors"
	"net/http"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary Get user levels
// @Description Return levels for the current user's target language, using the code when no native-language translation exists
// @ID user-list-levels
// @Tags user
// @Produce json
// @Success 200 {array} response.UserLevel
// @Failure 400,401,404,500 {object} response.Error
// @Security BearerAuth
// @Router /user/levels [get]
func (r *V1) listUserLevels(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	user, err := r.u.GetUser(ctx.UserContext(), userID)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "user not found")
		}
		r.l.Error(err, "restapi - v1 - user levels - get user")
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
	if user.NativeLanguageID == nil || user.TargetLanguageID == nil {
		return errorResponse(ctx, http.StatusBadRequest, "user languages are not configured")
	}
	levels, err := r.levels.ListLevels(ctx.UserContext(), user.TargetLanguageID)
	if err != nil {
		r.l.Error(err, "restapi - v1 - user levels - list levels")
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
	return ctx.Status(http.StatusOK).JSON(userLevelsResponse(levels, *user.NativeLanguageID))
}

func userLevelsResponse(levels []entity.Level, nativeLanguageID int) []response.UserLevel {
	result := make([]response.UserLevel, 0, len(levels))
	for _, level := range levels {
		name := level.Code
		for _, translation := range level.Translations {
			if translation.LanguageID == nativeLanguageID {
				name = translation.Name
				break
			}
		}
		result = append(result, response.UserLevel{ID: level.ID, Code: level.Code, Name: name})
	}
	return result
}
