package request

type TopicTranslation struct {
	LanguageID int    `json:"languageId" validate:"required,gt=0" example:"1"`
	Name       string `json:"name" validate:"required,max=100" example:"Travel"`
}

type SaveTopic struct {
	Slug         string             `json:"slug" validate:"required,max=100" example:"travel"`
	IconURL      string             `json:"iconUrl" example:"https://example.com/travel.svg"`
	IsActive     *bool              `json:"isActive" validate:"required" example:"true"`
	Translations []TopicTranslation `json:"translations" validate:"required,min=1,dive"`
}

type TopicStatus struct {
	IsActive *bool `json:"isActive" validate:"required" example:"true"`
}
