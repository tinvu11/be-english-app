package v1

import (
	"errors"
	"net/http"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/request"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary Translate or reuse a dictionary word
// @Description Uses the user's target language as source and native language as target; an existing dictionary entry is reused before DeepSeek is called
// @ID user-translate-vocabulary
// @Tags user-vocabulary
// @Accept json
// @Produce json
// @Param request body request.TranslateVocabulary true "Word to translate"
// @Success 200 {object} entity.VocabularyLookupResult
// @Failure 400,401,409,429,502 {object} map[string]string
// @Security BearerAuth
// @Router /vocabulary/translate [post]
func (r *V1) translateVocabulary(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok || userID == "" {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	var body request.TranslateVocabulary
	if err := ctx.BodyParser(&body); err != nil || r.v.Struct(body) != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid vocabulary")
	}
	result, err := r.vocabulary.TranslateWord(ctx.UserContext(), userID, body.Word)
	if err == nil {
		return ctx.JSON(result)
	}
	switch {
	case errors.Is(err, entity.ErrTargetLanguageRequired), errors.Is(err, entity.ErrNativeLanguageRequired):
		return errorResponse(ctx, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrInvalidVocabulary), errors.Is(err, entity.ErrInvalidLanguage):
		return errorResponse(ctx, http.StatusBadRequest, err.Error())
	case errors.Is(err, entity.ErrTranslationRateLimited):
		return errorResponse(ctx, http.StatusTooManyRequests, "translation rate limit reached")
	case errors.Is(err, entity.ErrTranslationUnavailable), errors.Is(err, entity.ErrTranslationFailed),
		errors.Is(err, entity.ErrInvalidVocabularyTranslation):
		return errorResponse(ctx, http.StatusBadGateway, "vocabulary translation failed")
	default:
		r.l.Error(err, "restapi - v1 - translate vocabulary")
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}
