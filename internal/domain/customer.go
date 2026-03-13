package domain

import (
	"errors"
	"strings"
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

//go:generate mockery --name=CustomerService --with-expecter
//go:generate mockery --name=CustomerRepository --with-expecter

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
	if body.Type == IndividualCustomerType {
		cpf := sanitizeCPF(body.Document)
		if err := validateCPF(cpf); err != nil {
			return Customer{}, BadRequestError{Message: err.Error()}
		}

		id := uuid.New().String()
		return Customer{
			ID:     id,
			UserID: id,
			Type:   body.Type,
			CPF:    cpf,
			Phone:  body.Phone,
			User: &User{
				ID:       id,
				Name:     body.Name,
				Email:    body.Email,
				Password: body.Password,
			},
		}, nil
	}

	if body.Type == CompanyCustomerType {
		if strings.TrimSpace(body.CompanyName) == "" {
			return Customer{}, BadRequestError{Message: "company_name is required for COMPANY type"}
		}

		cnpj := sanitizeCNPJ(body.Document)
		if err := validateCNPJ(cnpj); err != nil {
			return Customer{}, BadRequestError{Message: err.Error()}
		}

		id := uuid.New().String()
		return Customer{
			ID:          id,
			UserID:      id,
			Type:        body.Type,
			CNPJ:        cnpj,
			CompanyName: body.CompanyName,
			Phone:       body.Phone,
			User: &User{
				ID:       id,
				Name:     body.Name,
				Email:    body.Email,
				Password: body.Password,
			},
		}, nil
	}

	return Customer{}, BadRequestError{Message: "invalid customer type"}
}

// sanitizeCPF removes all non-digit characters (e.g. '.', '-') from a CPF string.
func sanitizeCPF(doc string) string {
	result := make([]rune, 0, len(doc))
	for _, r := range doc {
		if unicode.IsDigit(r) {
			result = append(result, r)
		}
	}
	return string(result)
}

// sanitizeCNPJ removes formatting characters ('.', '-', '/') and uppercases the result.
// Letters are preserved to support the alphanumeric CNPJ format (RFB 2243/2024).
func sanitizeCNPJ(doc string) string {
	upper := strings.ToUpper(doc)
	result := make([]rune, 0, len(upper))
	for _, r := range upper {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result = append(result, r)
		}
	}
	return string(result)
}

// validateCPF validates a Brazilian CPF.
// Rules: 11 digits, no all-same sequence, two check digits via weighted sum algorithm.
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
		return errors.New(invalidCPFErrorMessage)
	}
	if calcDigit(cpf, 10) != int(cpf[10]-'0') {
		return errors.New(invalidCPFErrorMessage)
	}

	return nil
}

// validateCNPJ validates a Brazilian CNPJ, supporting both numeric and alphanumeric formats.
// Alphanumeric format: characters A-Z have values 10-35 (RFB Instruction 2243/2024).
// Check digits (positions 13-14) must always be numeric.
func validateCNPJ(cnpj string) error {
	if len(cnpj) != 14 {
		return errors.New("CNPJ must have 14 characters")
	}

	// Check digits must be numeric
	if cnpj[12] < '0' || cnpj[12] > '9' || cnpj[13] < '0' || cnpj[13] > '9' {
		return errors.New(invalidCNPJErrorMessage)
	}

	// All characters must be valid alphanumeric (0-9 or A-Z)
	for _, c := range cnpj {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z')) {
			return errors.New(invalidCNPJErrorMessage)
		}
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

	cnpjCharValue := func(c byte) int {
		if c >= '0' && c <= '9' {
			return int(c - '0')
		}
		return int(c-'A') + 10
	}

	calcDigit := func(s string, length int, weights []int) int {
		sum := 0
		for i := 0; i < length; i++ {
			sum += cnpjCharValue(s[i]) * weights[i]
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
