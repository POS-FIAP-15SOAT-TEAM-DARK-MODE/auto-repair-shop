package customer

import "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"

type customerResponse struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Email       string              `json:"email"`
	Type        domain.CustomerType `json:"type"`
	Document    string              `json:"document"`
	CompanyName string              `json:"company_name,omitempty"`
	Phone       string              `json:"phone"`
}

func toResponse(c domain.Customer) customerResponse {
	var document string
	if c.Type == domain.IndividualCustomerType {
		document = c.CPF
	} else {
		document = c.CNPJ
	}

	return customerResponse{
		ID:          c.ID,
		Name:        c.User.Name,
		Email:       c.User.Email,
		Type:        c.Type,
		Document:    document,
		CompanyName: c.CompanyName,
		Phone:       c.Phone,
	}
}
