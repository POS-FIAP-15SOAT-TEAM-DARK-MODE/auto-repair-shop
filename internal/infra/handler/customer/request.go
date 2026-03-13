package customer

import "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"

type CreateCustomerRequest struct {
	Name        string `json:"name" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=6"`
	Type        string `json:"type" validate:"required,oneof=INDIVIDUAL COMPANY"`
	Document    string `json:"document" validate:"required"`
	CompanyName string `json:"company_name,omitempty"`
	Phone       string `json:"phone" validate:"required"`
}

func (r CreateCustomerRequest) toDomain() (domain.Customer, error) {
	return domain.CreateCustomerToDomain(r.Name, r.Email, r.Password, r.Type, r.Document, r.CompanyName, r.Phone)
}
