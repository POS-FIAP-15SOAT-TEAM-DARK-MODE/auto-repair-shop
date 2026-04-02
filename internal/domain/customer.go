package domain

import (
	"context"
	"errors"
	"strings"
	"unicode"
)

type CustomerType string

const (
	IndividualCustomerType CustomerType = "INDIVIDUAL"
	CompanyCustomerType    CustomerType = "COMPANY"

	cpfLength  = 11
	cnpjLength = 14
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=CustomerService --with-expecter
//go:generate go run github.com/vektra/mockery/v2@latest --name=CustomerRepository --with-expecter
type (
	Customer struct {
		ID          string
		UserID      string
		Type        CustomerType
		CPF         string
		CNPJ        string
		CompanyName string
		Phone       string
		User        *User
	}

	CustomerService interface {
		Create(ctx context.Context, customer Customer) error
		GetByID(ctx context.Context, id string) (Customer, error)
		GetByDocument(ctx context.Context, rawDocument string) (Customer, error)
		Update(ctx context.Context, id string, name, email, phone *string) (Customer, error)
		Delete(ctx context.Context, id string) error
	}

	CustomerRepository interface {
		Create(ctx context.Context, customer *Customer) error
		GetByID(ctx context.Context, id string) (Customer, error)
		GetByDocument(ctx context.Context, document string) (Customer, error)
		Update(ctx context.Context, id, phone string) error
		Delete(ctx context.Context, id string) error
	}
)

func NewCustomer(name, email, password string, customerType CustomerType, document, companyName, phone string) (Customer, error) {
	user, err := CreateUserToDomain(name, email, password)
	if err != nil {
		return Customer{}, err
	}

	c := Customer{
		ID:          user.ID,
		UserID:      user.ID,
		Type:        customerType,
		CompanyName: companyName,
		Phone:       phone,
		User:        user,
	}

	c.applyDocument(document)

	if err := c.Validate(); err != nil {
		return Customer{}, err
	}

	return c, nil
}

func (c *Customer) applyDocument(document string) {
	switch c.Type {
	case IndividualCustomerType:
		c.CPF = sanitizeCPF(document)
	case CompanyCustomerType:
		c.CNPJ = sanitizeCNPJ(document)
	}
}

func (c *Customer) Validate() error {
	var errs []error

	if strings.TrimSpace(c.Phone) == "" {
		errs = append(errs, ErrPhoneRequired)
	}

	switch c.Type {
	case IndividualCustomerType:
		if err := validateCPF(c.CPF); err != nil {
			errs = append(errs, err)
		}
	case CompanyCustomerType:
		if strings.TrimSpace(c.CompanyName) == "" {
			errs = append(errs, ErrCompanyNameRequired)
		}
		if err := validateCNPJ(c.CNPJ); err != nil {
			errs = append(errs, err)
		}
	default:
		errs = append(errs, ErrInvalidCustomerType)
	}

	return errors.Join(errs...)
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
		return ErrCPFLength
	}

	allSame := true
	for i := 1; i < 11; i++ {
		if cpf[i] != cpf[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return ErrInvalidCPF
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
		return ErrInvalidCPF
	}
	if calcDigit(cpf, 10) != int(cpf[10]-'0') {
		return ErrInvalidCPF
	}

	return nil
}

// validateCNPJ validates a Brazilian CNPJ, supporting both numeric and alphanumeric formats.
// Alphanumeric format: characters A-Z have values 10-35 (RFB Instruction 2243/2024).
// Check digits (positions 13-14) must always be numeric.
func validateCNPJ(cnpj string) error {
	if len(cnpj) != 14 {
		return ErrCNPJLength
	}

	// Check digits must be numeric
	if cnpj[12] < '0' || cnpj[12] > '9' || cnpj[13] < '0' || cnpj[13] > '9' {
		return ErrInvalidCNPJ
	}

	// All characters must be valid alphanumeric (0-9 or A-Z)
	for _, c := range cnpj {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z')) {
			return ErrInvalidCNPJ
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
		return ErrInvalidCNPJ
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
		return ErrInvalidCNPJ
	}
	if calcDigit(cnpj, 13, weights2) != int(cnpj[13]-'0') {
		return ErrInvalidCNPJ
	}

	return nil
}

// ParseDocument sanitizes a raw document string, determines whether it is a CPF or CNPJ,
// validates it using the existing check-digit algorithms, and returns the type and sanitized value.
func ParseDocument(raw string) (CustomerType, string, error) {
	// CNPJ sanitizer strips all non-alphanumeric chars and uppercases — safe for CPF too.
	sanitized := sanitizeCNPJ(raw)
	switch len(sanitized) {
	case cpfLength:
		if err := validateCPF(sanitized); err != nil {
			return "", "", err
		}
		return IndividualCustomerType, sanitized, nil
	case cnpjLength:
		if err := validateCNPJ(sanitized); err != nil {
			return "", "", err
		}
		return CompanyCustomerType, sanitized, nil
	default:
		return "", "", ErrInvalidDocumentFormat
	}
}
