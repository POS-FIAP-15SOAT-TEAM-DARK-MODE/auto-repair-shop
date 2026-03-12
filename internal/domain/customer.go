package domain

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/customer/dto"
)

const (
	individualCustomerType = "INDIVIDUAL"
	companyCustomerType    = "COMPANY"
)

type (
	CustomerService interface {
		Create(body dto.CreateCustomerRequest) (Customer, error)
	}

	CustomerRepository interface {
	}

	user struct {
		ID       string
		Name     string
		Email    string
		Password string
	}

	Customer struct {
		user
		ID          string
		userID      string
		Type        string
		CPF         string
		CNPJ        string
		CompanyName string
		Phone       string
	}
)

func (c *Customer) ToResponse() dto.CreateCustomerResponse {
	var document string

	if c.Type == individualCustomerType {
		document = c.CPF
	}

	if c.Type == companyCustomerType {
		document = c.CNPJ
	}

	return dto.CreateCustomerResponse{
		ID:          c.ID,
		Name:        c.Name,
		Email:       c.Email,
		Type:        c.Type,
		Document:    document,
		CompanyName: c.CompanyName,
		Phone:       c.Phone,
	}
}

func CreateCustomerToDomain(body dto.CreateCustomerRequest) Customer {
	customer := Customer{
		user: user{
			Name:     body.Name,
			Email:    body.Email,
			Password: body.Password,
		},
		Type:        body.Type,
		CompanyName: body.CompanyName,
		Phone:       body.Phone,
	}

	if body.Type == individualCustomerType {
		customer.CPF = body.Document
	}

	if body.Type == companyCustomerType {
		customer.CNPJ = body.Document
	}

	return customer
}
