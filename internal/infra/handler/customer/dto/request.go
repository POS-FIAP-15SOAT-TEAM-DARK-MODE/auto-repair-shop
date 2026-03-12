package dto

type (
	CreateCustomerRequest struct {
		Name        string `json:"name" validate:"required"`
		Email       string `json:"email" validate:"required,email"`
		Password    string `json:"password" validate:"required,min=6"`
		Type        string `json:"type" validate:"required,oneof=INDIVIDUAL COMPANY"`
		Document    string `json:"document" validate:"required"`
		CompanyName string `json:"company_name,omitempty"`
		Phone       string `json:"phone" validate:"required"`
	}
)
