package v1

import (
	"errors"
	"net/http"
	"strconv"

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

func vocabularyUserID(ctx *fiber.Ctx) (string, error) {
	userID, ok := ctx.Locals("userID").(string)
	if !ok || userID == "" {
		return "", entity.ErrUserNotFound
	}
	return userID, nil
}

func positivePathID(ctx *fiber.Ctx, name string) (int64, error) {
	id, err := strconv.ParseInt(ctx.Params(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, entity.ErrInvalidVocabulary
	}
	return id, nil
}

// @Summary Get vocabulary screen overview
// @Description Returns word totals and vocabulary categories for the current user's active language pair
// @Tags user-vocabulary
// @Produce json
// @Success 200 {object} entity.VocabularyOverview
// @Failure 401,409,500 {object} map[string]string
// @Security BearerAuth
// @Router /vocabulary/overview [get]
func (r *V1) getVocabularyOverview(ctx *fiber.Ctx) error {
	userID, err := vocabularyUserID(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	overview, err := r.vocabulary.GetOverview(ctx.UserContext(), userID)
	if err != nil {
		return r.vocabularyError(ctx, err, "get overview")
	}
	return ctx.JSON(overview)
}

// @Summary Create vocabulary set
// @Tags user-vocabulary
// @Accept json
// @Produce json
// @Param request body request.SaveVocabularySet true "Vocabulary set"
// @Success 201 {object} entity.VocabularySet
// @Security BearerAuth
// @Router /vocabulary/sets [post]
func (r *V1) createVocabularySet(ctx *fiber.Ctx) error {
	userID, err := vocabularyUserID(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	var body request.SaveVocabularySet
	if err = ctx.BodyParser(&body); err != nil || r.v.Struct(body) != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid vocabulary set")
	}
	item, err := r.vocabulary.CreateSet(ctx.UserContext(), userID, body.Title, body.ColorHex)
	if err != nil {
		return r.vocabularyError(ctx, err, "create set")
	}
	return ctx.Status(http.StatusCreated).JSON(item)
}

// @Summary List vocabulary sets
// @Tags user-vocabulary
// @Produce json
// @Security BearerAuth
// @Router /vocabulary/sets [get]
func (r *V1) listVocabularySets(ctx *fiber.Ctx) error {
	userID, err := vocabularyUserID(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	items, err := r.vocabulary.ListSets(ctx.UserContext(), userID)
	if err != nil {
		return r.vocabularyError(ctx, err, "list sets")
	}
	return ctx.JSON(fiber.Map{"items": items, "total": len(items)})
}

// @Summary Update vocabulary set
// @Tags user-vocabulary
// @Accept json
// @Produce json
// @Param setId path int true "Set ID"
// @Param request body request.SaveVocabularySet true "Vocabulary set"
// @Success 200 {object} entity.VocabularySet
// @Security BearerAuth
// @Router /vocabulary/sets/{setId} [put]
func (r *V1) updateVocabularySet(ctx *fiber.Ctx) error {
	userID, err := vocabularyUserID(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	id, err := positivePathID(ctx, "setId")
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid vocabulary set id")
	}
	var body request.SaveVocabularySet
	if err = ctx.BodyParser(&body); err != nil || r.v.Struct(body) != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid vocabulary set")
	}
	item, err := r.vocabulary.UpdateSet(ctx.UserContext(), userID, id, body.Title, body.ColorHex)
	if err != nil {
		return r.vocabularyError(ctx, err, "update set")
	}
	return ctx.JSON(item)
}

// @Summary Delete vocabulary set and its saved words
// @Tags user-vocabulary
// @Param setId path int true "Set ID"
// @Success 204
// @Security BearerAuth
// @Router /vocabulary/sets/{setId} [delete]
func (r *V1) deleteVocabularySet(ctx *fiber.Ctx) error {
	userID, err := vocabularyUserID(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	id, err := positivePathID(ctx, "setId")
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid vocabulary set id")
	}
	if err = r.vocabulary.DeleteSet(ctx.UserContext(), userID, id); err != nil {
		return r.vocabularyError(ctx, err, "delete set")
	}
	return ctx.SendStatus(http.StatusNoContent)
}

// @Summary Save vocabulary for current user
// @Tags user-vocabulary
// @Accept json
// @Produce json
// @Param request body request.CreateUserVocabulary true "Saved vocabulary"
// @Success 201 {object} entity.UserVocabulary
// @Security BearerAuth
// @Router /vocabulary/words [post]
func (r *V1) createUserVocabulary(ctx *fiber.Ctx) error {
	userID, err := vocabularyUserID(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	var body request.CreateUserVocabulary
	if err = ctx.BodyParser(&body); err != nil || r.v.Struct(body) != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid vocabulary")
	}
	item, err := r.vocabulary.CreateUserVocabulary(ctx.UserContext(), userID, entity.UserVocabularyInput{
		DictionaryID: body.DictionaryID, VocabSetID: body.VocabSetID, CaptionID: body.CaptionID})
	if err != nil {
		return r.vocabularyError(ctx, err, "create word")
	}
	return ctx.Status(http.StatusCreated).JSON(item)
}

// @Summary List current user's vocabulary
// @Tags user-vocabulary
// @Produce json
// @Param vocabSetId query int false "Filter by set ID"
// @Param search query string false "Search by word (case-insensitive)" maxlength(150)
// @Security BearerAuth
// @Router /vocabulary/words [get]
func (r *V1) listUserVocabularies(ctx *fiber.Ctx) error {
	userID, err := vocabularyUserID(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	var setID *int64
	if raw := ctx.Query("vocabSetId"); raw != "" {
		id, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil || id <= 0 {
			return errorResponse(ctx, http.StatusBadRequest, "invalid vocabulary set id")
		}
		setID = &id
	}
	items, err := r.vocabulary.ListUserVocabularies(ctx.UserContext(), userID, entity.UserVocabularyFilter{
		VocabSetID: setID,
		Search:     ctx.Query("search"),
	})
	if err != nil {
		return r.vocabularyError(ctx, err, "list words")
	}
	return ctx.JSON(fiber.Map{"items": items, "total": len(items)})
}

// @Summary List unlearned vocabulary
// @Description Returns all unlearned words when limit is omitted; when limit is present, returns a random selection of up to that size
// @Tags user-vocabulary
// @Produce json
// @Param vocabSetId query int false "Filter by set ID"
// @Param limit query int false "Random word count (1-1000)" minimum(1) maximum(1000)
// @Success 200 {object} map[string]interface{}
// @Failure 400,401,409,500 {object} map[string]string
// @Security BearerAuth
// @Router /vocabulary/words/unlearned [get]
func (r *V1) listUnlearnedVocabularies(ctx *fiber.Ctx) error {
	userID, err := vocabularyUserID(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	filter := entity.UnlearnedVocabularyFilter{}
	if raw := ctx.Query("vocabSetId"); raw != "" {
		id, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil || id <= 0 {
			return errorResponse(ctx, http.StatusBadRequest, "invalid vocabulary set id")
		}
		filter.VocabSetID = &id
	}
	if raw := ctx.Query("limit"); raw != "" {
		limit, parseErr := strconv.Atoi(raw)
		if parseErr != nil || limit <= 0 || limit > 1000 {
			return errorResponse(ctx, http.StatusBadRequest, "invalid limit")
		}
		filter.Limit = &limit
	}
	items, err := r.vocabulary.ListUnlearnedVocabularies(ctx.UserContext(), userID, filter)
	if err != nil {
		return r.vocabularyError(ctx, err, "list unlearned words")
	}
	return ctx.JSON(fiber.Map{"items": items, "total": len(items)})
}

// @Summary Update saved vocabulary
// @Description Move a word to a set, change its source caption, or mark it learned
// @Tags user-vocabulary
// @Accept json
// @Produce json
// @Param wordId path int true "Saved vocabulary ID"
// @Param request body request.UpdateUserVocabulary true "Vocabulary update"
// @Success 200 {object} entity.UserVocabulary
// @Security BearerAuth
// @Router /vocabulary/words/{wordId} [put]
func (r *V1) updateUserVocabulary(ctx *fiber.Ctx) error {
	userID, err := vocabularyUserID(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	id, err := positivePathID(ctx, "wordId")
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid vocabulary id")
	}
	var body request.UpdateUserVocabulary
	if err = ctx.BodyParser(&body); err != nil || r.v.Struct(body) != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid vocabulary")
	}
	item, err := r.vocabulary.UpdateUserVocabulary(ctx.UserContext(), userID, id, entity.UserVocabularyUpdate{
		VocabSetID: body.VocabSetID, CaptionID: body.CaptionID, IsLearned: body.IsLearned})
	if err != nil {
		return r.vocabularyError(ctx, err, "update word")
	}
	return ctx.JSON(item)
}

// @Summary Delete saved vocabulary
// @Tags user-vocabulary
// @Param wordId path int true "Saved vocabulary ID"
// @Success 204
// @Security BearerAuth
// @Router /vocabulary/words/{wordId} [delete]
func (r *V1) deleteUserVocabulary(ctx *fiber.Ctx) error {
	userID, err := vocabularyUserID(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	id, err := positivePathID(ctx, "wordId")
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid vocabulary id")
	}
	if err = r.vocabulary.DeleteUserVocabulary(ctx.UserContext(), userID, id); err != nil {
		return r.vocabularyError(ctx, err, "delete word")
	}
	return ctx.SendStatus(http.StatusNoContent)
}

func (r *V1) vocabularyError(ctx *fiber.Ctx, err error, operation string) error {
	switch {
	case errors.Is(err, entity.ErrInvalidVocabulary), errors.Is(err, entity.ErrInvalidVocabularySet):
		return errorResponse(ctx, http.StatusBadRequest, err.Error())
	case errors.Is(err, entity.ErrTargetLanguageRequired), errors.Is(err, entity.ErrNativeLanguageRequired):
		return errorResponse(ctx, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrVocabularySetNotFound), errors.Is(err, entity.ErrUserVocabularyNotFound), errors.Is(err, entity.ErrDictionaryNotFound):
		return errorResponse(ctx, http.StatusNotFound, err.Error())
	case errors.Is(err, entity.ErrUserVocabularyExists), errors.Is(err, entity.ErrVocabularyLanguageMismatch):
		return errorResponse(ctx, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrInvalidReference):
		return errorResponse(ctx, http.StatusBadRequest, "invalid dictionary, vocabulary set, or caption reference")
	default:
		r.l.Error(err, "restapi - v1 - vocabulary - "+operation)
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}
