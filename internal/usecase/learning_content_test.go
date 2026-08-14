package usecase_test

import (
	"context"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/usecase/learningcontent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type learningGeneratorStub struct {
	quizCalls      int
	localizedCalls int
}

func (stub *learningGeneratorStub) GenerateQuizzes(context.Context, entity.QuizGenerationRequest) ([]entity.GeneratedQuiz, error) {
	stub.quizCalls++
	return nil, nil
}

func (stub *learningGeneratorStub) GenerateLocalizedContent(context.Context, entity.LocalizedContentGenerationRequest) (entity.GeneratedLocalizedContent, error) {
	stub.localizedCalls++
	return entity.GeneratedLocalizedContent{Summary: "Tóm tắt", Vocabulary: []entity.GeneratedVocabulary{{
		Word: "environment", Meaning: "môi trường", Example1Sentence: "Example one.",
		Example1Translation: "Ví dụ một.", Example2Sentence: "Example two.", Example2Translation: "Ví dụ hai.",
	}}}, nil
}

func TestLearningContentReusesQuizForNewNativeLanguage(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	content := NewMockLearningContentRepo(ctrl)
	dictionary := NewMockDictionaryRepo(ctrl)
	captions := NewMockCaptionRepo(ctrl)
	videos := NewMockVideoRepo(ctrl)
	users := NewMockUserRepo(ctrl)
	languages := NewMockLanguageRepo(ctrl)
	generator := &learningGeneratorStub{}
	nativeID := 2
	video := entity.Video{ID: 10, Status: entity.VideoStatusPublished, Language: entity.VideoLanguage{ID: 1, Code: "en", Name: "English"}}
	target := entity.Language{ID: nativeID, Code: "vi", Name: "Vietnamese", IsActive: true}
	quiz := entity.LearningQuiz{ID: 5, Question: "Question?", CorrectOption: 0,
		Options: []entity.LearningQuizOption{{Order: 0, Content: "A"}, {Order: 1, Content: "B"}, {Order: 2, Content: "C"}, {Order: 3, Content: "D"}}}
	entry := entity.DictionaryEntry{ID: 7, Word: "environment", SourceLanguageID: 1, TargetLanguageID: nativeID, Meaning: "môi trường"}

	users.EXPECT().GetByID(gomock.Any(), "user-1").Return(entity.User{NativeLanguageID: &nativeID}, nil)
	videos.EXPECT().GetVideo(gomock.Any(), int64(10)).Return(video, nil)
	languages.EXPECT().GetLanguage(gomock.Any(), nativeID).Return(target, nil)
	content.EXPECT().GetQuizzes(gomock.Any(), int64(10)).Return([]entity.LearningQuiz{quiz}, nil).Times(2)
	content.EXPECT().GetLocalizedLearningContent(gomock.Any(), int64(10), nativeID).
		Return(entity.VideoSummary{}, nil, entity.ErrLearningContentNotFound)
	captions.EXPECT().ListCaptions(gomock.Any(), int64(10), entity.CaptionFilter{Limit: 10000}).
		Return(entity.CaptionList{Items: []entity.Caption{{SentenceOrder: 1, Content: "Protect the environment."}}}, nil)
	dictionary.EXPECT().UpsertDictionaryEntry(gomock.Any(), gomock.Any()).Return(entry, true, nil)
	content.EXPECT().SaveLocalizedLearningContentIfAbsent(gomock.Any(), gomock.Any(), []int64{7}).Return(nil)
	content.EXPECT().GetLocalizedLearningContent(gomock.Any(), int64(10), nativeID).
		Return(entity.VideoSummary{VideoID: 10, LanguageID: nativeID, Content: "Tóm tắt"}, []entity.DictionaryEntry{entry}, nil)

	result, err := learningcontent.New(content, dictionary, captions, videos, users, languages, generator).
		GetForUser(t.Context(), "user-1", 10)
	require.NoError(t, err)
	assert.Equal(t, "Tóm tắt", result.Summary)
	assert.Equal(t, 0, generator.quizCalls)
	assert.Equal(t, 1, generator.localizedCalls)
}
