package v1

import (
	"net/http"
	"strconv"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary     Get onboarding data
// @Description Return active learning languages, translation languages, and active topics. Pass translationLanguageId to localize topic names; without it, the first available translation is used.
// @ID          user-onboarding
// @Tags        user
// @Produce     json
// @Param       translationLanguageId query int false "Language ID used for topic names"
// @Success     200 {object} response.Onboarding
// @Failure     400 {object} response.Error
// @Failure     401 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /user/onboarding [get]
func (r *V1) onboarding(ctx *fiber.Ctx) error {
	translationLanguageID, err := optionalPositiveInt(ctx.Query("translationLanguageId"))
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid translation language id")
	}

	languages, err := r.languages.ListLanguages(ctx.UserContext())
	if err != nil {
		r.l.Error(err, "restapi - v1 - onboarding - list languages")
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
	topics, err := r.topics.ListTopics(ctx.UserContext())
	if err != nil {
		r.l.Error(err, "restapi - v1 - onboarding - list topics")
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(http.StatusOK).JSON(buildOnboardingResponse(languages, topics, translationLanguageID))
}

func optionalPositiveInt(raw string) (*int, error) {
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return nil, entity.ErrInvalidLanguage
	}
	return &value, nil
}

func buildOnboardingResponse(languages []entity.Language, topics []entity.Topic, translationLanguageID *int) response.Onboarding {
	result := response.Onboarding{
		LearningLanguages:    make([]response.OnboardingLanguage, 0),
		TranslationLanguages: make([]response.OnboardingLanguage, 0),
		Topics:               make([]response.OnboardingTopic, 0),
	}
	for _, language := range languages {
		if !language.IsActive {
			continue
		}
		item := response.OnboardingLanguage{ID: language.ID, Code: language.Code, Name: language.Name, Flag: language.FlagEmoji}
		result.TranslationLanguages = append(result.TranslationLanguages, item)
		if language.IsLearnable {
			result.LearningLanguages = append(result.LearningLanguages, item)
		}
	}
	for _, topic := range topics {
		if !topic.IsActive {
			continue
		}
		name := topicName(topic.Translations, translationLanguageID)
		if name == "" {
			continue
		}
		result.Topics = append(result.Topics, response.OnboardingTopic{ID: topic.ID, Slug: topic.Slug, Icon: topic.IconURL, Name: name})
	}
	return result
}

func topicName(translations []entity.TopicTranslation, languageID *int) string {
	if languageID == nil {
		if len(translations) == 0 {
			return ""
		}
		return translations[0].Name
	}
	for _, translation := range translations {
		if translation.LanguageID == *languageID {
			return translation.Name
		}
	}
	return ""
}
