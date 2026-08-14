package entity

import "time"

type DictionaryEntry struct {
	ID                  int64     `json:"id"`
	Word                string    `json:"word"`
	SourceLanguageID    int       `json:"sourceLanguageId"`
	TargetLanguageID    int       `json:"targetLanguageId"`
	PhoneticOrPinyin    string    `json:"phoneticOrPinyin,omitempty"`
	PartOfSpeech        string    `json:"partOfSpeech,omitempty"`
	Meaning             string    `json:"meaning"`
	Example1Sentence    string    `json:"example1Sentence,omitempty"`
	Example1Translation string    `json:"example1Translation,omitempty"`
	Example2Sentence    string    `json:"example2Sentence,omitempty"`
	Example2Translation string    `json:"example2Translation,omitempty"`
	CreatedAt           time.Time `json:"createdAt"`
}

type DictionaryInput struct {
	Word                string
	SourceLanguageID    int
	TargetLanguageID    int
	PhoneticOrPinyin    string
	PartOfSpeech        string
	Meaning             string
	Example1Sentence    string
	Example1Translation string
	Example2Sentence    string
	Example2Translation string
}

type VocabularyTranslationRequest struct {
	Word               string `json:"word"`
	SourceLanguageCode string `json:"sourceLanguageCode"`
	SourceLanguageName string `json:"sourceLanguageName"`
	TargetLanguageCode string `json:"targetLanguageCode"`
	TargetLanguageName string `json:"targetLanguageName"`
}

type VocabularyTranslation struct {
	PhoneticOrPinyin    string `json:"phoneticOrPinyin"`
	PartOfSpeech        string `json:"partOfSpeech"`
	Meaning             string `json:"meaning"`
	Example1Sentence    string `json:"example1Sentence"`
	Example1Translation string `json:"example1Translation"`
	Example2Sentence    string `json:"example2Sentence"`
	Example2Translation string `json:"example2Translation"`
}

type VocabularyLookupResult struct {
	Entry  DictionaryEntry `json:"entry"`
	Reused bool            `json:"reused"`
}

type VocabularySet struct {
	ID               int64     `json:"id" example:"1"`
	Title            string    `json:"title" example:"Travel"`
	SourceLanguageID int       `json:"sourceLanguageId" example:"1"`
	TargetLanguageID int       `json:"targetLanguageId" example:"2"`
	WordCount        int       `json:"wordCount" example:"20"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type VocabularyLanguages struct {
	SourceLanguageID int
	TargetLanguageID int
}

type VocabularyCategorySummary struct {
	ID             int64  `json:"id" example:"1"`
	Title          string `json:"title" example:"Travel"`
	TotalWords     int    `json:"totalWords" example:"20"`
	LearnedWords   int    `json:"learnedWords" example:"12"`
	UnlearnedWords int    `json:"unlearnedWords" example:"8"`
}

type VocabularyOverview struct {
	TotalWords     int                         `json:"totalWords" example:"50"`
	LearnedWords   int                         `json:"learnedWords" example:"30"`
	UnlearnedWords int                         `json:"unlearnedWords" example:"20"`
	Categories     []VocabularyCategorySummary `json:"categories"`
}

type UserVocabulary struct {
	ID         int64           `json:"id" example:"1"`
	VocabSetID int64           `json:"vocabSetId" example:"2"`
	CaptionID  *int64          `json:"captionId,omitempty"`
	Entry      DictionaryEntry `json:"entry"`
	IsLearned  bool            `json:"isLearned" example:"false"`
	LearnedAt  *time.Time      `json:"learnedAt,omitempty"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

type UserVocabularyInput struct {
	DictionaryID int64
	VocabSetID   int64
	CaptionID    *int64
}

type UserVocabularyUpdate struct {
	VocabSetID int64
	CaptionID  *int64
	IsLearned  bool
}
