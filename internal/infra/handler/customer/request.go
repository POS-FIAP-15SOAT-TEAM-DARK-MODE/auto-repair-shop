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
		return ErrDocumentRequired
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

type getCustomerByIDRequest struct {
	ID string
}

func (r *getCustomerByIDRequest) validate() error {
	if strings.TrimSpace(r.ID) == "" {
		return ErrIDRequired
	}
	return nil
}

type getCustomerByDocumentRequest struct {
	Document string
}

func (r *getCustomerByDocumentRequest) validate() error {
	if strings.TrimSpace(r.Document) == "" {
		return ErrDocumentQueryRequired
	}
	return nil
}

type updateCustomerRequest struct {
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty"`
	Phone *string `json:"phone,omitempty"`
}

func (r *updateCustomerRequest) validate() error {
	if r.Name != nil {
		u := &domain.User{Name: *r.Name}
		if err := u.IsValidName(); err != nil {
			return err
		}
	}
	if r.Email != nil {
		u := &domain.User{Email: *r.Email}
		if err := u.IsValidEmail(); err != nil {
			return err
		}
	}
	if r.Phone != nil && strings.TrimSpace(*r.Phone) == "" {
		return domain.ErrPhoneRequired
	}
	return nil
}
