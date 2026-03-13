package domain

import (
	"errors"
	"unicode"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/customer/dto"
	"github.com/google/uuid"
)

const (
	IndividualCustomerType  = "INDIVIDUAL"
	CompanyCustomerType     = "COMPANY"
	invalidCPFErrorMessage  = "invalid CPF"
	invalidCNPJErrorMessage = "invalid CNPJ"
)

type (
	Customer struct {
		ID          string
		UserID      string
		Type        string
		CPF         string
		CNPJ        string
		CompanyName string
		Phone       string
		User        *User
	}

	CustomerService interface {
		Create(body dto.CreateCustomerRequest) (Customer, error)
	}

	CustomerRepository interface {
		Create(customer *Customer) error
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

func CreateCustomerToDomain(body dto.CreateCustomerRequest) (Customer, error) {
	doc := sanitizeDocument(body.Document)

	if body.Type == IndividualCustomerType {
		if err := validateCPF(doc); err != nil {
			return Customer{}, err
		}
	}

	if body.Type == CompanyCustomerType {
		if err := validateCNPJ(doc); err != nil {
			return Customer{}, err
		}
	}

	id := uuid.New().String()

	customer := Customer{
		ID:          id,
		UserID:      id,
		Type:        body.Type,
		CompanyName: body.CompanyName,
		Phone:       body.Phone,
		User: &User{
			ID:       id,
			Name:     body.Name,
			Email:    body.Email,
			Password: body.Password,
		},
	}

	if body.Type == IndividualCustomerType {
		customer.CPF = doc
	}

	if body.Type == CompanyCustomerType {
		customer.CNPJ = doc
	}

	return customer, nil
}

func sanitizeDocument(doc string) string {
	result := make([]rune, 0, len(doc))

	for _, r := range doc {
		if unicode.IsDigit(r) {
			result = append(result, r)
		}
	}

	return string(result)
}

func validateCPF(cpf string) error {
	if len(cpf) != 11 {
		return errors.New("CPF must have 11 digits")
	}

	allSame := true
	for i := 1; i < 11; i++ {
		if cpf[i] != cpf[0] {
			allSame = false
			break
		}
	}

	if allSame {
		return errors.New(invalidCPFErrorMessage)
	}

	calcDigit := func(s string, length int) int {
		sum := 0

		for i := 0; i < length; i++ {
			sum += int(s[i]-'0') * (length + 1 - i)
		}

		remainder := (sum * 10) % 11
		if remainder == 10 {
			return 0
		}

		return remainder
	}

	if calcDigit(cpf, 9) != int(cpf[9]-'0') {
		return errors.New("invalid CPF")
	}
	if calcDigit(cpf, 10) != int(cpf[10]-'0') {
		return errors.New("invalid CPF")
	}

	return nil
}

func validateCNPJ(cnpj string) error {
	if len(cnpj) != 14 {
		return errors.New("CNPJ must have 14 digits")
	}

	allSame := true
	for i := 1; i < 14; i++ {
		if cnpj[i] != cnpj[0] {
			allSame = false
			break
		}
	}

	if allSame {
		return errors.New(invalidCNPJErrorMessage)
	}

	calcDigit := func(s string, length int, weights []int) int {
		sum := 0
		for i := 0; i < length; i++ {
			sum += int(s[i]-'0') * weights[i]
		}
		remainder := sum % 11
		if remainder < 2 {
			return 0
		}
		return 11 - remainder
	}

	weights1 := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	weights2 := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}

	if calcDigit(cnpj, 12, weights1) != int(cnpj[12]-'0') {
		return errors.New(invalidCNPJErrorMessage)
	}
	if calcDigit(cnpj, 13, weights2) != int(cnpj[13]-'0') {
		return errors.New(invalidCNPJErrorMessage)
	}

	return nil
}
