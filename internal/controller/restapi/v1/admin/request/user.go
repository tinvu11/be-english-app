package request

type UserStatus struct {
	IsActive *bool `json:"isActive" validate:"required" example:"false"`
}

type UserRole struct {
	Role string `json:"role" validate:"required,oneof=user admin" example:"admin"`
}
