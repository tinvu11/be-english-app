package entity

import "time"

type TopicTranslation struct {
	LanguageID int    `json:"languageId" example:"1"`
	Name       string `json:"name" example:"Travel"`
} // @name entity.TopicTranslation

type Topic struct {
	ID           int                `json:"id" example:"1"`
	Slug         string             `json:"slug" example:"travel"`
	IconURL      string             `json:"iconUrl,omitempty" example:"https://example.com/travel.svg"`
	IsActive     bool               `json:"isActive" example:"true"`
	CreatedAt    time.Time          `json:"createdAt" example:"2026-08-06T00:00:00Z"`
	Translations []TopicTranslation `json:"translations"`
} // @name entity.Topic
