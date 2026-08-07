package admin

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	adminrequest "github.com/evrone/go-clean-template/internal/controller/restapi/v1/admin/request"
	_ "github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

const maxSRTFileSize = 10 << 20

// @Summary List all captions of a video
// @Description Sorted by sentenceOrder ascending and includes all translations
// @Tags admin-captions
// @Produce json
// @Param videoId path int true "Video ID"
// @Success 200 {array} entity.Caption
// @Failure 400,401,403,404,500 {object} response.Error
// @Security BearerAuth
// @Router /admin/videos/{videoId}/captions [get]
func (ctrl *controller) listCaptions(ctx *fiber.Ctx) error {
	videoID, err := positiveInt64(ctx.Params("videoId"))
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video id")
	}
	items, err := ctrl.captions.ListCaptions(ctx.UserContext(), videoID)
	if err != nil {
		return ctrl.captionError(ctx, err)
	}
	return ctx.JSON(items)
}

// @Summary Import original and translated SRT files
// @Description Use original_file once; repeat translation_files and translation_language_ids in matching order
// @Tags admin-captions
// @Accept multipart/form-data
// @Produce json
// @Param videoId path int true "Video ID"
// @Param original_file formData file true "Original SRT"
// @Param translation_files formData file false "Translated SRT files"
// @Param translation_language_ids formData []int false "Language IDs matching translation files"
// @Success 201 {array} entity.Caption
// @Failure 400,401,403,404,409,413,500 {object} response.Error
// @Security BearerAuth
// @Router /admin/videos/{videoId}/captions/import [post]
func (ctrl *controller) importCaptions(ctx *fiber.Ctx) error {
	videoID, err := positiveInt64(ctx.Params("videoId"))
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video id")
	}
	form, err := ctx.MultipartForm()
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid multipart form")
	}
	originalHeaders := form.File["original_file"]
	if len(originalHeaders) != 1 {
		return errorResponse(ctx, http.StatusBadRequest, "original_file is required")
	}
	original, err := readSRTFile(originalHeaders[0])
	if err != nil {
		return ctrl.srtFileError(ctx, err)
	}
	translationHeaders := form.File["translation_files"]
	languageValues := form.Value["translation_language_ids"]
	if len(translationHeaders) != len(languageValues) {
		return errorResponse(ctx, http.StatusBadRequest, "translation files and language ids must have the same count")
	}
	translations := make([]entity.SRTTranslationFile, len(translationHeaders))
	for index, header := range translationHeaders {
		languageID, parseErr := strconv.Atoi(languageValues[index])
		if parseErr != nil || languageID <= 0 {
			return errorResponse(ctx, http.StatusBadRequest, "invalid translation language id")
		}
		data, readErr := readSRTFile(header)
		if readErr != nil {
			return ctrl.srtFileError(ctx, readErr)
		}
		translations[index] = entity.SRTTranslationFile{LanguageID: languageID, Data: data}
	}
	items, err := ctrl.captions.ImportSRT(ctx.UserContext(), videoID, original, translations)
	if err != nil {
		ctrl.log.Warn("admin caption SRT import rejected: video_id=%d original=%q translations=%d error=%v",
			videoID, originalHeaders[0].Filename, len(translations), err)
		return ctrl.captionError(ctx, err)
	}
	return ctx.Status(http.StatusCreated).JSON(items)
}

// @Summary Add one caption manually
// @Tags admin-captions
// @Accept json
// @Produce json
// @Param videoId path int true "Video ID"
// @Param request body adminrequest.SaveCaption true "Caption"
// @Success 201 {object} entity.Caption
// @Failure 400,401,403,404,409,500 {object} response.Error
// @Security BearerAuth
// @Router /admin/videos/{videoId}/captions [post]
func (ctrl *controller) createCaption(ctx *fiber.Ctx) error {
	videoID, err := positiveInt64(ctx.Params("videoId"))
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video id")
	}
	body, err := ctrl.parseCaption(ctx)
	if err != nil {
		return err
	}
	item, err := ctrl.captions.CreateCaption(ctx.UserContext(), videoID, captionFromRequest(body))
	if err != nil {
		return ctrl.captionError(ctx, err)
	}
	return ctx.Status(http.StatusCreated).JSON(item)
}

// @Summary Update a caption and all translations
// @Tags admin-captions
// @Accept json
// @Produce json
// @Param videoId path int true "Video ID"
// @Param captionId path int true "Caption ID"
// @Param request body adminrequest.SaveCaption true "Caption"
// @Success 200 {object} entity.Caption
// @Failure 400,401,403,404,409,500 {object} response.Error
// @Security BearerAuth
// @Router /admin/videos/{videoId}/captions/{captionId} [put]
func (ctrl *controller) updateCaption(ctx *fiber.Ctx) error {
	videoID, captionID, err := captionIDs(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video or caption id")
	}
	body, err := ctrl.parseCaption(ctx)
	if err != nil {
		return err
	}
	item, err := ctrl.captions.UpdateCaption(ctx.UserContext(), videoID, captionID, captionFromRequest(body))
	if err != nil {
		return ctrl.captionError(ctx, err)
	}
	return ctx.JSON(item)
}

// @Summary Delete one caption
// @Tags admin-captions
// @Param videoId path int true "Video ID"
// @Param captionId path int true "Caption ID"
// @Success 204
// @Failure 400,401,403,404,500 {object} response.Error
// @Security BearerAuth
// @Router /admin/videos/{videoId}/captions/{captionId} [delete]
func (ctrl *controller) deleteCaption(ctx *fiber.Ctx) error {
	videoID, captionID, err := captionIDs(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid video or caption id")
	}
	if err = ctrl.captions.DeleteCaption(ctx.UserContext(), videoID, captionID); err != nil {
		return ctrl.captionError(ctx, err)
	}
	return ctx.SendStatus(http.StatusNoContent)
}

func (ctrl *controller) parseCaption(ctx *fiber.Ctx) (adminrequest.SaveCaption, error) {
	var body adminrequest.SaveCaption
	if err := ctx.BodyParser(&body); err != nil {
		return body, errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}
	if err := ctrl.validate.Struct(body); err != nil || body.EndTimeMS < body.StartTimeMS {
		return body, errorResponse(ctx, http.StatusBadRequest, "invalid caption")
	}
	return body, nil
}

func captionFromRequest(body adminrequest.SaveCaption) entity.CaptionInput {
	translations := make([]entity.CaptionTranslationInput, len(body.Translations))
	for index, item := range body.Translations {
		translations[index] = entity.CaptionTranslationInput{LanguageID: item.LanguageID, Text: item.Text}
	}
	return entity.CaptionInput{SentenceOrder: body.SentenceOrder, StartTimeMS: body.StartTimeMS,
		EndTimeMS: body.EndTimeMS, Content: body.Content, PinyinOrFurigana: body.PinyinOrFurigana,
		Translations: translations}
}

func captionIDs(ctx *fiber.Ctx) (int64, int64, error) {
	videoID, err := positiveInt64(ctx.Params("videoId"))
	if err != nil {
		return 0, 0, err
	}
	captionID, err := positiveInt64(ctx.Params("captionId"))
	return videoID, captionID, err
}

func readSRTFile(header *multipart.FileHeader) ([]byte, error) {
	if !strings.EqualFold(filepath.Ext(header.Filename), ".srt") {
		return nil, entity.ErrInvalidSRT
	}
	if header.Size > maxSRTFileSize {
		return nil, fiber.ErrRequestEntityTooLarge
	}
	file, err := header.Open()
	if err != nil {
		return nil, entity.ErrInvalidSRT
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxSRTFileSize+1))
	if err != nil {
		return nil, entity.ErrInvalidSRT
	}
	if len(data) > maxSRTFileSize {
		return nil, fiber.ErrRequestEntityTooLarge
	}
	return data, nil
}

func (ctrl *controller) srtFileError(ctx *fiber.Ctx, err error) error {
	if errors.Is(err, fiber.ErrRequestEntityTooLarge) {
		return errorResponse(ctx, http.StatusRequestEntityTooLarge, "SRT file is too large")
	}
	return errorResponse(ctx, http.StatusBadRequest, "invalid SRT file")
}

func (ctrl *controller) captionError(ctx *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, entity.ErrInvalidCaption), errors.Is(err, entity.ErrInvalidReference), errors.Is(err, entity.ErrInvalidSRT):
		return errorResponse(ctx, http.StatusBadRequest, err.Error())
	case errors.Is(err, entity.ErrCaptionTranslationCount):
		return errorResponse(ctx, http.StatusBadRequest, "translation caption count does not match original")
	case errors.Is(err, entity.ErrVideoNotFound):
		return errorResponse(ctx, http.StatusNotFound, "video not found")
	case errors.Is(err, entity.ErrCaptionNotFound):
		return errorResponse(ctx, http.StatusNotFound, "caption not found")
	case errors.Is(err, entity.ErrCaptionExists):
		return errorResponse(ctx, http.StatusConflict, "caption sentence order already exists")
	default:
		ctrl.log.Error(err, "restapi - v1 - admin - caption")
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}
