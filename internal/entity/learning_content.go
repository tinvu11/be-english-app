package entity

import "time"

type LearningCaption struct {
	Order int    `json:"order"`
	Text  string `json:"text"`
}

type LearningQuizOption struct {
	Order   int    `json:"order"`
	Content string `json:"content"`
}

type LearningQuiz struct {
	ID            int64                `json:"id"`
	Order         int                  `json:"order"`
	Question      string               `json:"question"`
	Options       []LearningQuizOption `json:"options"`
	CorrectOption int                  `json:"correctOption"`
	Explanation   string               `json:"explanation"`
}

type GeneratedQuiz struct {
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	CorrectOption int      `json:"correctOption"`
	Explanation   string   `json:"explanation"`
}

type QuizGenerationRequest struct {
	SourceLanguageCode string            `json:"sourceLanguageCode"`
	Captions           []LearningCaption `json:"captions"`
}

type GeneratedVocabulary struct {
	Word                string `json:"word"`
	PhoneticOrPinyin    string `json:"phoneticOrPinyin"`
	PartOfSpeech        string `json:"partOfSpeech"`
	Meaning             string `json:"meaning"`
	Example1Sentence    string `json:"example1Sentence"`
	Example1Translation string `json:"example1Translation"`
	Example2Sentence    string `json:"example2Sentence"`
	Example2Translation string `json:"example2Translation"`
}

type LocalizedContentGenerationRequest struct {
	SourceLanguageCode string            `json:"sourceLanguageCode"`
	SourceLanguageName string            `json:"sourceLanguageName"`
	TargetLanguageCode string            `json:"targetLanguageCode"`
	TargetLanguageName string            `json:"targetLanguageName"`
	Captions           []LearningCaption `json:"captions"`
}

type GeneratedLocalizedContent struct {
	Summary    string                `json:"summary"`
	Vocabulary []GeneratedVocabulary `json:"vocabulary"`
}

type VideoSummary struct {
	VideoID      int64     `json:"videoId"`
	LanguageID   int       `json:"languageId"`
	LanguageCode string    `json:"languageCode"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"createdAt"`
}

type VideoLearningContent struct {
	VideoID      int64             `json:"videoId"`
	LanguageID   int               `json:"languageId"`
	LanguageCode string            `json:"languageCode"`
	Summary      string            `json:"summary"`
	Quizzes      []LearningQuiz    `json:"quizzes"`
	Vocabulary   []DictionaryEntry `json:"vocabulary"`
}
