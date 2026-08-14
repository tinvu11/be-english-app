package learningcontent

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/gateway"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/internal/usecase"
)

const maxCaptions = 10000

type UseCase struct {
	content    repo.LearningContentRepo
	dictionary repo.DictionaryRepo
	captions   repo.CaptionRepo
	videos     repo.VideoRepo
	users      repo.UserRepo
	languages  repo.LanguageRepo
	generator  gateway.LearningContentGenerator
}

func New(content repo.LearningContentRepo, dictionary repo.DictionaryRepo, captions repo.CaptionRepo, videos repo.VideoRepo,
	users repo.UserRepo, languages repo.LanguageRepo, generator gateway.LearningContentGenerator) usecase.LearningContent {
	return &UseCase{content: content, dictionary: dictionary, captions: captions, videos: videos,
		users: users, languages: languages, generator: generator}
}

func (uc *UseCase) GetForUser(ctx context.Context, userID string, videoID int64) (entity.VideoLearningContent, error) {
	if strings.TrimSpace(userID) == "" || videoID <= 0 {
		return entity.VideoLearningContent{}, entity.ErrInvalidVideo
	}
	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return entity.VideoLearningContent{}, err
	}
	if user.NativeLanguageID == nil || *user.NativeLanguageID <= 0 {
		return entity.VideoLearningContent{}, entity.ErrNativeLanguageRequired
	}
	video, err := uc.videos.GetVideo(ctx, videoID)
	if err != nil {
		return entity.VideoLearningContent{}, err
	}
	if video.Status != entity.VideoStatusPublished {
		return entity.VideoLearningContent{}, entity.ErrVideoNotFound
	}
	target, err := uc.languages.GetLanguage(ctx, *user.NativeLanguageID)
	if err != nil {
		return entity.VideoLearningContent{}, err
	}
	if !target.IsActive || video.Language.ID <= 0 {
		return entity.VideoLearningContent{}, entity.ErrInvalidLanguage
	}
	quizzes, quizErr := uc.content.GetQuizzes(ctx, videoID)
	summary, vocabulary, localizedErr := uc.content.GetLocalizedLearningContent(ctx, videoID, target.ID)
	if quizErr == nil && localizedErr == nil {
		return buildResult(videoID, target, summary.Content, quizzes, vocabulary), nil
	}
	if (quizErr != nil && !errors.Is(quizErr, entity.ErrLearningContentNotFound)) ||
		(localizedErr != nil && !errors.Is(localizedErr, entity.ErrLearningContentNotFound)) {
		if quizErr != nil && !errors.Is(quizErr, entity.ErrLearningContentNotFound) {
			return entity.VideoLearningContent{}, quizErr
		}
		return entity.VideoLearningContent{}, localizedErr
	}
	captionList, err := uc.captions.ListCaptions(ctx, videoID, entity.CaptionFilter{Limit: maxCaptions})
	if err != nil {
		return entity.VideoLearningContent{}, err
	}
	if len(captionList.Items) == 0 {
		return entity.VideoLearningContent{}, entity.ErrCaptionTranslationEmpty
	}
	captionInput := make([]entity.LearningCaption, 0, len(captionList.Items))
	for _, caption := range captionList.Items {
		captionInput = append(captionInput, entity.LearningCaption{Order: caption.SentenceOrder, Text: caption.Content})
	}
	if errors.Is(quizErr, entity.ErrLearningContentNotFound) {
		generated, generateErr := uc.generator.GenerateQuizzes(ctx, entity.QuizGenerationRequest{
			SourceLanguageCode: video.Language.Code, Captions: captionInput})
		if generateErr != nil {
			return entity.VideoLearningContent{}, generateErr
		}
		if !validQuizzes(generated) {
			return entity.VideoLearningContent{}, entity.ErrInvalidLearningContent
		}
		if err = uc.content.SaveQuizzesIfAbsent(ctx, videoID, generated); err != nil {
			return entity.VideoLearningContent{}, err
		}
	}
	if errors.Is(localizedErr, entity.ErrLearningContentNotFound) {
		generated, generateErr := uc.generator.GenerateLocalizedContent(ctx, entity.LocalizedContentGenerationRequest{
			SourceLanguageCode: video.Language.Code, SourceLanguageName: video.Language.Name,
			TargetLanguageCode: target.Code, TargetLanguageName: target.Name, Captions: captionInput})
		if generateErr != nil {
			return entity.VideoLearningContent{}, generateErr
		}
		if !validLocalized(generated) {
			return entity.VideoLearningContent{}, entity.ErrInvalidLearningContent
		}
		dictionaryIDs := make([]int64, 0, len(generated.Vocabulary))
		seenWords := make(map[string]struct{}, len(generated.Vocabulary))
		for _, item := range generated.Vocabulary {
			word := strings.ToLower(strings.TrimSpace(item.Word))
			if _, exists := seenWords[word]; exists {
				continue
			}
			seenWords[word] = struct{}{}
			entry, _, upsertErr := uc.dictionary.UpsertDictionaryEntry(ctx, entity.DictionaryInput{
				Word: word, SourceLanguageID: video.Language.ID, TargetLanguageID: target.ID,
				PhoneticOrPinyin: strings.TrimSpace(item.PhoneticOrPinyin), PartOfSpeech: strings.TrimSpace(item.PartOfSpeech),
				Meaning: strings.TrimSpace(item.Meaning), Example1Sentence: strings.TrimSpace(item.Example1Sentence),
				Example1Translation: strings.TrimSpace(item.Example1Translation), Example2Sentence: strings.TrimSpace(item.Example2Sentence),
				Example2Translation: strings.TrimSpace(item.Example2Translation)})
			if upsertErr != nil {
				return entity.VideoLearningContent{}, upsertErr
			}
			dictionaryIDs = append(dictionaryIDs, entry.ID)
		}
		if err = uc.content.SaveLocalizedLearningContentIfAbsent(ctx, entity.VideoSummary{
			VideoID: videoID, LanguageID: target.ID, Content: strings.TrimSpace(generated.Summary)}, dictionaryIDs); err != nil {
			return entity.VideoLearningContent{}, err
		}
	}
	quizzes, err = uc.content.GetQuizzes(ctx, videoID)
	if err != nil {
		return entity.VideoLearningContent{}, err
	}
	summary, vocabulary, err = uc.content.GetLocalizedLearningContent(ctx, videoID, target.ID)
	if err != nil {
		return entity.VideoLearningContent{}, err
	}
	return buildResult(videoID, target, summary.Content, quizzes, vocabulary), nil
}

func validQuizzes(items []entity.GeneratedQuiz) bool {
	if len(items) == 0 || len(items) > 20 {
		return false
	}
	for _, item := range items {
		if strings.TrimSpace(item.Question) == "" || strings.TrimSpace(item.Explanation) == "" ||
			len(item.Options) != 4 || item.CorrectOption < 0 || item.CorrectOption >= len(item.Options) {
			return false
		}
		for _, option := range item.Options {
			if strings.TrimSpace(option) == "" {
				return false
			}
		}
	}
	return true
}

func validLocalized(item entity.GeneratedLocalizedContent) bool {
	if strings.TrimSpace(item.Summary) == "" || len(item.Vocabulary) == 0 || len(item.Vocabulary) > 50 {
		return false
	}
	for _, word := range item.Vocabulary {
		if strings.TrimSpace(word.Word) == "" || utf8.RuneCountInString(strings.TrimSpace(word.Word)) > 150 ||
			strings.TrimSpace(word.Meaning) == "" || strings.TrimSpace(word.Example1Sentence) == "" ||
			strings.TrimSpace(word.Example1Translation) == "" || strings.TrimSpace(word.Example2Sentence) == "" ||
			strings.TrimSpace(word.Example2Translation) == "" {
			return false
		}
	}
	return true
}

func buildResult(videoID int64, target entity.Language, summary string, quizzes []entity.LearningQuiz, vocabulary []entity.DictionaryEntry) entity.VideoLearningContent {
	return entity.VideoLearningContent{VideoID: videoID, LanguageID: target.ID, LanguageCode: target.Code,
		Summary: summary, Quizzes: quizzes, Vocabulary: vocabulary}
}
