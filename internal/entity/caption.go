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

type YouTubeSubtitleTrack struct {
	LanguageCode string   `json:"languageCode" example:"en"`
	Name         string   `json:"name,omitempty" example:"English"`
	Formats      []string `json:"formats"`
}

type YouTubeSubtitleTracks struct {
	Tracks []YouTubeSubtitleTrack `json:"tracks"`
}

type YouTubeCaptionImportResult struct {
	VideoID       int64  `json:"videoId" example:"10"`
	LanguageCode  string `json:"languageCode" example:"en"`
	Source        string `json:"source" example:"youtube_manual"`
	ImportedCount int    `json:"importedCount" example:"128"`
}
