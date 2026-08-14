package entity

import "time"

type CaptionTranslation struct {
	LanguageID   int    `json:"languageId" example:"2"`
	LanguageCode string `json:"languageCode" example:"vi"`
	LanguageName string `json:"languageName" example:"Vietnamese"`
	Text         string `json:"text" example:"Xin chào"`
}

type Caption struct {
	ID               int64                `json:"id" example:"1"`
	VideoID          int64                `json:"videoId" example:"10"`
	SentenceOrder    int                  `json:"sentenceOrder" example:"1"`
	StartTimeMS      int64                `json:"startTimeMs" example:"1000"`
	EndTimeMS        int64                `json:"endTimeMs" example:"3500"`
	Content          string               `json:"content" example:"Hello"`
	PinyinOrFurigana string               `json:"pinyinOrFurigana,omitempty"`
	Translations     []CaptionTranslation `json:"translations"`
	CreatedAt        time.Time            `json:"createdAt"`
}

type CaptionFilter struct {
	Limit  int
	Offset int
}

type CaptionList struct {
	Items []Caption `json:"items"`
	Total int       `json:"total" example:"500"`
}

type CaptionTranslationInput struct {
	LanguageID int
	Text       string
}

type CaptionInput struct {
	SentenceOrder    int
	StartTimeMS      int64
	EndTimeMS        int64
	Content          string
	PinyinOrFurigana string
	Translations     []CaptionTranslationInput
}

type SRTTranslationFile struct {
	LanguageID int
	Data       []byte
}

const (
	CaptionImportFailIfExists = "fail_if_exists"
	CaptionImportReplaceAll   = "replace_all"
)

type YouTubeCaptionImportResult struct {
	VideoID       int64  `json:"videoId" example:"10"`
	LanguageCode  string `json:"languageCode" example:"en"`
	Source        string `json:"source" example:"groq_whisper"`
	ImportedCount int    `json:"importedCount" example:"128"`
}

type AudioTranscriptionSegment struct {
	StartSeconds float64
	EndSeconds   float64
	Text         string
}

type AudioTranscription struct {
	LanguageCode string
	Segments     []AudioTranscriptionSegment
}

const (
	CaptionTranslationMissingOnly = "missing_only"
	CaptionTranslationReplace     = "replace"
)

type CaptionText struct {
	CaptionID int64  `json:"captionId"`
	Order     int    `json:"order"`
	Text      string `json:"text"`
}

type CaptionTranslationRequest struct {
	SourceLanguage string        `json:"sourceLanguage"`
	TargetLanguage string        `json:"targetLanguage"`
	Items          []CaptionText `json:"items"`
}

type TranslatedCaption struct {
	CaptionID int64  `json:"captionId"`
	Order     int    `json:"order"`
	Text      string `json:"text"`
}

type CaptionTranslationUpsert struct {
	CaptionID int64
	Text      string
}

type CaptionTranslationResult struct {
	VideoID            int64  `json:"videoId"`
	SourceLanguageCode string `json:"sourceLanguageCode"`
	TargetLanguageCode string `json:"targetLanguageCode"`
	TranslatedCount    int    `json:"translatedCount"`
	SkippedCount       int    `json:"skippedCount"`
}

type VideoCaptionItem struct {
	ID            int64  `json:"id"`
	SentenceOrder int    `json:"sentenceOrder"`
	StartTimeMS   int64  `json:"startTimeMs"`
	EndTimeMS     int64  `json:"endTimeMs"`
	Text          string `json:"text"`
}

type VideoCaptions struct {
	VideoID      int64              `json:"videoId"`
	LanguageID   int                `json:"languageId"`
	LanguageCode string             `json:"languageCode"`
	Items        []VideoCaptionItem `json:"items"`
	Total        int                `json:"total"`
}
