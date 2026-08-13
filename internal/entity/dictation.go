package entity

import "time"

type DictationProgress struct {
	VideoID       int64     `json:"video_id" example:"1"`
	CaptionID     int64     `json:"caption_id" example:"10"`
	SentenceOrder int       `json:"sentence_order" example:"3"`
	StartTimeMS   int64     `json:"start_time_ms" example:"12000"`
	EndTimeMS     int64     `json:"end_time_ms" example:"15500"`
	Content       string    `json:"content" example:"This is a dictation sentence."`
	CompletedAt   time.Time `json:"completed_at" example:"2026-08-12T10:30:00Z"`
} // @name entity.DictationProgress

type DictationProgressList struct {
	Items []DictationProgress `json:"items"`
	Total int                 `json:"total" example:"12"`
} // @name entity.DictationProgressList
