package request

type TranslateVocabulary struct {
	Word string `json:"word" validate:"required,max=150" example:"remember"`
}

type SaveVocabularySet struct {
	Title string `json:"title" validate:"required,max=255" example:"Travel"`
}

type CreateUserVocabulary struct {
	DictionaryID int64  `json:"dictionaryId" validate:"required,gt=0" example:"12"`
	VocabSetID   int64  `json:"vocabSetId" validate:"required,gt=0" example:"2"`
	CaptionID    *int64 `json:"captionId,omitempty" validate:"omitempty,gt=0" example:"35"`
}

type UpdateUserVocabulary struct {
	VocabSetID int64  `json:"vocabSetId" validate:"required,gt=0" example:"2"`
	CaptionID  *int64 `json:"captionId,omitempty" validate:"omitempty,gt=0" example:"35"`
	IsLearned  bool   `json:"isLearned" example:"true"`
}
