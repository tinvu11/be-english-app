package entity

import (
	"encoding/json"
	"time"
)

const ShadowingProviderAzure = "azure"

type PronunciationAssessmentInput struct {
	Audio         []byte
	ContentType   string
	ReferenceText string
	Locale        string
}

type PronunciationAssessment struct {
	Provider           string
	AccuracyScore      float64
	FluencyScore       float64
	CompletenessScore  float64
	PronunciationScore float64
	ProsodyScore       *float64
	DurationMS         int64
	Words              []ShadowingWordResult
	ProviderResponse   json.RawMessage
}

type ShadowingPrompt struct {
	VideoID       int64
	CaptionID     int64
	ReferenceText string
	LanguageCode  string
}

type ShadowingWordResult struct {
	Order         int             `json:"order" example:"0"`
	Word          string          `json:"word" example:"comfortable"`
	AccuracyScore float64         `json:"accuracy_score" example:"54"`
	ErrorType     string          `json:"error_type,omitempty" example:"Mispronunciation"`
	Phonemes      json.RawMessage `json:"phonemes,omitempty" swaggertype:"object"`
}

type ShadowingAttempt struct {
	ID                 int64                 `json:"id" example:"123"`
	VideoID            int64                 `json:"video_id" example:"1"`
	CaptionID          int64                 `json:"caption_id" example:"10"`
	Provider           string                `json:"provider" example:"azure"`
	Locale             string                `json:"locale" example:"en-US"`
	ReferenceText      string                `json:"reference_text" example:"Good morning."`
	AccuracyScore      float64               `json:"accuracy_score" example:"79"`
	FluencyScore       float64               `json:"fluency_score" example:"86"`
	CompletenessScore  float64               `json:"completeness_score" example:"100"`
	PronunciationScore float64               `json:"pronunciation_score" example:"82"`
	ProsodyScore       *float64              `json:"prosody_score,omitempty" example:"75"`
	DurationMS         int64                 `json:"duration_ms" example:"2300"`
	Passed             bool                  `json:"passed" example:"true"`
	Words              []ShadowingWordResult `json:"words"`
	CreatedAt          time.Time             `json:"created_at"`
	ProviderResponse   json.RawMessage       `json:"-"`
}

type ShadowingAttemptList struct {
	Items []ShadowingAttempt `json:"items"`
	Total int                `json:"total" example:"3"`
}
