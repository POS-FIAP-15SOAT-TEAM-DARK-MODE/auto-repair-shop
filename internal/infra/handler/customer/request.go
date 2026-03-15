package customer

import (
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type CreateCustomerRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Type        string `json:"type"`
	Document    string `json:"document"`
	CompanyName string `json:"company_name,omitempty"`
	Phone       string `json:"phone"`
}

// validate checks DTO-specific required fields that have no direct equivalent in the domain.
// document is validated here because the domain only knows CPF and CNPJ, not a raw document field.
func (r *CreateCustomerRequest) validate() error {
	if strings.TrimSpace(r.Document) == "" {
		return domain.ValidationError{Message: "document is required"}
	}
	return nil
}

func (r *CreateCustomerRequest) toDomain() (domain.Customer, error) {
	customer, err := domain.CreateCustomerToDomain(r.Name, r.Email, r.Password, r.Type, r.Document, r.CompanyName, r.Phone)
	if err != nil {
		return domain.Customer{}, domain.ValidationError{Message: err.Error()}
	}
	return customer, nil
}
