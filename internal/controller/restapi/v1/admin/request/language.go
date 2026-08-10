package request

type CreateLanguage struct {
	Code        string `json:"code" validate:"required,max=10" example:"en"`
	Name        string `json:"name" validate:"required,max=50" example:"English"`
	FlagEmoji   string `json:"flagEmoji" validate:"required,max=16" example:"🇬🇧"`
	IsLearnable bool   `json:"isLearnable" example:"false"`
}

type UpdateLanguage struct {
	Name        string `json:"name" validate:"required,max=50" example:"English"`
	FlagEmoji   string `json:"flagEmoji" validate:"required,max=16" example:"🇬🇧"`
	IsActive    *bool  `json:"isActive" validate:"required" example:"true"`
	IsLearnable bool   `json:"isLearnable" example:"false"`
}
