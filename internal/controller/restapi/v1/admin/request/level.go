package request

type LevelTranslation struct {
	LanguageID int    `json:"languageId" validate:"required,gt=0" example:"1"`
	Name       string `json:"name" validate:"required,max=100" example:"Beginner"`
}

type SaveLevel struct {
	Code         string             `json:"code" validate:"required,max=20" example:"A1"`
	LanguageID   int                `json:"languageId" validate:"required,gt=0" example:"1"`
	Translations []LevelTranslation `json:"translations" validate:"required,min=1,dive"`
}
