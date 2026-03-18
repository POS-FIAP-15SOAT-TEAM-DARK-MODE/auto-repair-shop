package customer

import (
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type createCustomerRequest struct {
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
func (r *createCustomerRequest) validate() error {
	if strings.TrimSpace(r.Document) == "" {
		return domain.ValidationError{Message: "document is required"}
	}
	return nil
}

func (r *createCustomerRequest) Domain() (domain.Customer, error) {
	customer, err := domain.NewCustomer(r.Name, r.Email, r.Password, domain.CustomerType(r.Type), r.Document, r.CompanyName, r.Phone)
	if err != nil {
		return domain.Customer{}, domain.ValidationError{Message: err.Error()}
	}
	return customer, nil
}
