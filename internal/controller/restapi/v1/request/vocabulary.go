package request

type TranslateVocabulary struct {
	Word string `json:"word" validate:"required,max=150" example:"remember"`
}
