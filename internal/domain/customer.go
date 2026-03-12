package domain

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/customer/dto"
)

const (
	IndividualCustomerType = "INDIVIDUAL"
	CompanyCustomerType    = "COMPANY"
)

type (
	Customer struct {
		ID          string
		UserID      string
		User        *User
		Type        string
		CPF         string
		CNPJ        string
		CompanyName string
		Phone       string
	}

	CustomerService interface {
		Create(body dto.CreateCustomerRequest) (Customer, error)
	}

	CustomerRepository interface {
		Create(customer Customer) error
	}
)

func (c *Customer) ToResponse() dto.CreateCustomerResponse {
	var document string

	if c.Type == IndividualCustomerType {
		document = c.CPF
	}

	if c.Type == CompanyCustomerType {
		document = c.CNPJ
	}

	return dto.CreateCustomerResponse{
		ID:          c.ID,
		Name:        c.User.Name,
		Email:       c.User.Email,
		Type:        c.Type,
		Document:    document,
		CompanyName: c.CompanyName,
		Phone:       c.Phone,
	}
}

func CreateCustomerToDomain(body dto.CreateCustomerRequest) Customer {
	customer := Customer{
		User: &User{
			Name:     body.Name,
			Email:    body.Email,
			Password: body.Password,
		},
		Type:        body.Type,
		CompanyName: body.CompanyName,
		Phone:       body.Phone,
	}

	if body.Type == IndividualCustomerType {
		customer.CPF = body.Document
	}

	if body.Type == CompanyCustomerType {
		customer.CNPJ = body.Document
	}

	return customer
}
