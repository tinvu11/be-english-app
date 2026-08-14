package usecase_test

import (
	"context"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/usecase/caption"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type translatorStub struct{}

type transcriberStub struct{}

func (transcriberStub) Transcribe(context.Context, []byte, string) (entity.AudioTranscription, error) {
	return entity.AudioTranscription{}, nil
}

func (translatorStub) TranslateCaptions(_ context.Context, input entity.CaptionTranslationRequest) ([]entity.TranslatedCaption, error) {
	result := make([]entity.TranslatedCaption, len(input.Items))
	for index, item := range input.Items {
		result[index] = entity.TranslatedCaption{CaptionID: item.CaptionID, Order: item.Order, Text: "Xin chao"}
	}
	return result, nil
}

func TestGetTranslatedCaptionsTranslatesAndReturnsNativeLanguage(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	captions := NewMockCaptionRepo(ctrl)
	videos := NewMockVideoRepo(ctrl)
	languages := NewMockLanguageRepo(ctrl)
	users := NewMockUserRepo(ctrl)
	nativeLanguageID := 2
	videoEntity := entity.Video{ID: 10, Status: entity.VideoStatusPublished, Language: entity.VideoLanguage{ID: 1, Code: "en"}}
	target := entity.Language{ID: nativeLanguageID, Code: "vi", IsActive: true}
	sourceCaption := entity.Caption{ID: 5, VideoID: 10, SentenceOrder: 1, StartTimeMS: 100,
		EndTimeMS: 900, Content: "Hello"}
	translatedCaption := sourceCaption
	translatedCaption.Translations = []entity.CaptionTranslation{{LanguageID: nativeLanguageID, Text: "Xin chao"}}

	users.EXPECT().GetByID(gomock.Any(), "user-1").Return(entity.User{NativeLanguageID: &nativeLanguageID}, nil)
	videos.EXPECT().GetVideo(gomock.Any(), int64(10)).Return(videoEntity, nil).Times(2)
	languages.EXPECT().GetLanguage(gomock.Any(), nativeLanguageID).Return(target, nil).Times(2)
	captions.EXPECT().ListCaptionsForTranslation(gomock.Any(), int64(10), nativeLanguageID, true).
		Return([]entity.Caption{sourceCaption}, 1, nil)
	captions.EXPECT().UpsertTranslations(gomock.Any(), int64(10), nativeLanguageID,
		[]entity.CaptionTranslationUpsert{{CaptionID: 5, Text: "Xin chao"}}).Return(nil)
	captions.EXPECT().ListCaptions(gomock.Any(), int64(10), entity.CaptionFilter{Limit: 10000, Offset: 0}).
		Return(entity.CaptionList{Items: []entity.Caption{translatedCaption}, Total: 1}, nil)

	result, err := caption.New(captions, videos, languages, users, youtubeProviderStub{}, transcriberStub{}, translatorStub{}, 50).
		GetTranslatedCaptions(t.Context(), "user-1", 10)
	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assert.Equal(t, "vi", result.LanguageCode)
	assert.Equal(t, "Xin chao", result.Items[0].Text)
}

func TestGetOriginalCaptionsDoesNotLoadUserOrTranslate(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	captions := NewMockCaptionRepo(ctrl)
	videos := NewMockVideoRepo(ctrl)
	languages := NewMockLanguageRepo(ctrl)
	users := NewMockUserRepo(ctrl)
	videoEntity := entity.Video{ID: 10, Status: entity.VideoStatusPublished,
		Language: entity.VideoLanguage{ID: 1, Code: "en"}}
	source := entity.Caption{ID: 5, SentenceOrder: 1, StartTimeMS: 100, EndTimeMS: 900, Content: "Hello"}
	videos.EXPECT().GetVideo(gomock.Any(), int64(10)).Return(videoEntity, nil)
	captions.EXPECT().ListCaptions(gomock.Any(), int64(10), entity.CaptionFilter{Limit: 10000, Offset: 0}).
		Return(entity.CaptionList{Items: []entity.Caption{source}, Total: 1}, nil)

	result, err := caption.New(captions, videos, languages, users, youtubeProviderStub{}, transcriberStub{}, translatorStub{}, 50).
		GetOriginalCaptions(t.Context(), 10)
	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assert.Equal(t, "en", result.LanguageCode)
	assert.Equal(t, "Hello", result.Items[0].Text)
}
