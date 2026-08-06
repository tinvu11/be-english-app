package request

type CreateLanguage struct {
	Code string `json:"code" validate:"required,max=10" example:"en"`
	Name string `json:"name" validate:"required,max=50" example:"English"`
}

type UpdateLanguage struct {
	Name     string `json:"name" validate:"required,max=50" example:"English"`
	IsActive *bool  `json:"isActive" validate:"required" example:"true"`
}
