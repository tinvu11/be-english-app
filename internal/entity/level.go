package entity

import "time"

type LevelTranslation struct {
	LanguageID int    `json:"languageId" example:"1"`
	Name       string `json:"name" example:"Beginner"`
} // @name entity.LevelTranslation

type Level struct {
	ID           int                `json:"id" example:"1"`
	Code         string             `json:"code" example:"A1"`
	LanguageID   int                `json:"languageId" example:"1"`
	CreatedAt    time.Time          `json:"createdAt" example:"2026-08-06T00:00:00Z"`
	Translations []LevelTranslation `json:"translations"`
} // @name entity.Level
