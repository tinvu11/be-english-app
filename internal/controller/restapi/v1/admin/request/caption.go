package request

type CaptionTranslation struct {
	LanguageID int    `json:"languageId" validate:"required,gt=0" example:"2"`
	Text       string `json:"text" validate:"required" example:"Xin chào"`
}

type SaveCaption struct {
	SentenceOrder    int                  `json:"sentenceOrder" validate:"min=0" example:"1"`
	StartTimeMS      int64                `json:"startTimeMs" validate:"min=0" example:"1000"`
	EndTimeMS        int64                `json:"endTimeMs" validate:"min=0" example:"3500"`
	Content          string               `json:"content" validate:"required" example:"Hello"`
	PinyinOrFurigana string               `json:"pinyinOrFurigana" example:"Nǐ hǎo"`
	Translations     []CaptionTranslation `json:"translations" validate:"dive"`
}

type ImportYouTubeCaptions struct {
	LanguageCode string `json:"languageCode" validate:"required,max=35" example:"en"`
	Mode         string `json:"mode" validate:"required,oneof=fail_if_exists replace_all" example:"replace_all"`
}
