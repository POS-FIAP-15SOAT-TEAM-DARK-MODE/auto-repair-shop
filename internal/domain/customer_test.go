package domain_test

import (
	"strings"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/customer/dto"
)

// ── CPF ──────────────────────────────────────────────────────────────────────

func TestCreateCustomerToDomain_CPF_Valid(t *testing.T) {
	body := dto.CreateCustomerRequest{
		Name:     "João Silva",
		Email:    "joao@example.com",
		Password: "secret123",
		Type:     domain.IndividualCustomerType,
		Document: "111.444.777-35", // formatted — must be accepted
		Phone:    "11999999999",
	}

	customer, err := domain.CreateCustomerToDomain(body)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if customer.CPF != "11144477735" {
		t.Errorf("expected sanitized CPF '11144477735', got '%s'", customer.CPF)
	}
	if customer.ID == "" {
		t.Error("expected non-empty ID")
	}
	if customer.User == nil {
		t.Fatal("expected User to be set")
	}
	if customer.User.ID != customer.ID {
		t.Error("expected User.ID to equal Customer.ID")
	}
	if customer.CNPJ != "" || customer.CompanyName != "" {
		t.Error("INDIVIDUAL customer must not have CNPJ or CompanyName")
	}
}

func TestCreateCustomerToDomain_CPF_PlainDigits(t *testing.T) {
	body := dto.CreateCustomerRequest{
		Name:     "Maria",
		Email:    "maria@example.com",
		Password: "secret123",
		Type:     domain.IndividualCustomerType,
		Document: "11144477735",
		Phone:    "11999999999",
	}

	_, err := domain.CreateCustomerToDomain(body)
	if err != nil {
		t.Fatalf("expected no error for plain CPF, got: %v", err)
	}
}

func TestCreateCustomerToDomain_CPF_InvalidCheckDigit(t *testing.T) {
	body := dto.CreateCustomerRequest{
		Name:     "Test",
		Email:    "test@example.com",
		Password: "secret123",
		Type:     domain.IndividualCustomerType,
		Document: "111.444.777-36", // last digit wrong
		Phone:    "11999999999",
	}

	_, err := domain.CreateCustomerToDomain(body)
	if err == nil {
		t.Fatal("expected error for invalid CPF check digit")
	}

	var badReq domain.BadRequestError
	if !asBadRequest(err, &badReq) {
		t.Errorf("expected BadRequestError, got %T", err)
	}
}

func TestCreateCustomerToDomain_CPF_AllSameDigits(t *testing.T) {
	cases := []string{
		"111.111.111-11",
		"000.000.000-00",
		"99999999999",
	}

	for _, doc := range cases {
		body := dto.CreateCustomerRequest{
			Name: "Test", Email: "t@e.com", Password: "secret123",
			Type: domain.IndividualCustomerType, Document: doc, Phone: "11999999999",
		}
		_, err := domain.CreateCustomerToDomain(body)
		if err == nil {
			t.Errorf("expected error for all-same CPF %q", doc)
		}
	}
}

func TestCreateCustomerToDomain_CPF_WrongLength(t *testing.T) {
	cases := []string{"1234567890", "123456789012"} // 10 and 12 digits

	for _, doc := range cases {
		body := dto.CreateCustomerRequest{
			Name: "Test", Email: "t@e.com", Password: "secret123",
			Type: domain.IndividualCustomerType, Document: doc, Phone: "11999999999",
		}
		_, err := domain.CreateCustomerToDomain(body)
		if err == nil {
			t.Errorf("expected error for CPF with wrong length %q", doc)
		}
	}
}

// ── CNPJ (numeric) ───────────────────────────────────────────────────────────

func TestCreateCustomerToDomain_CNPJ_NumericValid(t *testing.T) {
	body := dto.CreateCustomerRequest{
		Name:        "Empresa SA",
		Email:       "empresa@example.com",
		Password:    "secret123",
		Type:        domain.CompanyCustomerType,
		Document:    "11.222.333/0001-81", // formatted
		CompanyName: "Empresa SA",
		Phone:       "11999999999",
	}

	customer, err := domain.CreateCustomerToDomain(body)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if customer.CNPJ != "11222333000181" {
		t.Errorf("expected sanitized CNPJ '11222333000181', got '%s'", customer.CNPJ)
	}
	if customer.CompanyName != "Empresa SA" {
		t.Errorf("expected company name 'Empresa SA', got '%s'", customer.CompanyName)
	}
	if customer.CPF != "" {
		t.Error("COMPANY customer must not have CPF")
	}
}

func TestCreateCustomerToDomain_CNPJ_NumericInvalidCheckDigit(t *testing.T) {
	body := dto.CreateCustomerRequest{
		Name: "Empresa", Email: "e@e.com", Password: "secret123",
		Type:        domain.CompanyCustomerType,
		Document:    "11.222.333/0001-82", // last digit wrong
		CompanyName: "Empresa",
		Phone:       "11999999999",
	}

	_, err := domain.CreateCustomerToDomain(body)
	if err == nil {
		t.Fatal("expected error for invalid CNPJ check digit")
	}
}

func TestCreateCustomerToDomain_CNPJ_AllSame(t *testing.T) {
	cases := []string{
		"11.111.111/1111-11",
		"00000000000000",
	}

	for _, doc := range cases {
		body := dto.CreateCustomerRequest{
			Name: "Test", Email: "t@e.com", Password: "secret123",
			Type: domain.CompanyCustomerType, Document: doc,
			CompanyName: "Test", Phone: "11999999999",
		}
		_, err := domain.CreateCustomerToDomain(body)
		if err == nil {
			t.Errorf("expected error for all-same CNPJ %q", doc)
		}
	}
}

// ── CNPJ (alphanumeric — RFB 2243/2024) ──────────────────────────────────────

func TestCreateCustomerToDomain_CNPJ_AlphanumericValid(t *testing.T) {
	// ABCDEFGH000195 — check digits: d1=9, d2=5 (pre-calculated)
	body := dto.CreateCustomerRequest{
		Name:        "Tech Ltda",
		Email:       "tech@example.com",
		Password:    "secret123",
		Type:        domain.CompanyCustomerType,
		Document:    "AB.CDE.FGH/0001-95", // formatted alphanumeric
		CompanyName: "Tech Ltda",
		Phone:       "11999999999",
	}

	customer, err := domain.CreateCustomerToDomain(body)

	if err != nil {
		t.Fatalf("expected no error for alphanumeric CNPJ, got: %v", err)
	}
	if customer.CNPJ != "ABCDEFGH000195" {
		t.Errorf("expected 'ABCDEFGH000195', got '%s'", customer.CNPJ)
	}
}

func TestCreateCustomerToDomain_CNPJ_AlphanumericLowercase(t *testing.T) {
	// Lowercase input must be accepted and normalized to uppercase
	body := dto.CreateCustomerRequest{
		Name:        "Tech Ltda",
		Email:       "tech@example.com",
		Password:    "secret123",
		Type:        domain.CompanyCustomerType,
		Document:    "ab.cde.fgh/0001-95",
		CompanyName: "Tech Ltda",
		Phone:       "11999999999",
	}

	customer, err := domain.CreateCustomerToDomain(body)

	if err != nil {
		t.Fatalf("expected no error for lowercase alphanumeric CNPJ, got: %v", err)
	}
	if customer.CNPJ != "ABCDEFGH000195" {
		t.Errorf("expected normalized 'ABCDEFGH000195', got '%s'", customer.CNPJ)
	}
}

func TestCreateCustomerToDomain_CNPJ_AlphanumericInvalidCheckDigit(t *testing.T) {
	// Change last digit to make it invalid
	body := dto.CreateCustomerRequest{
		Name: "Test", Email: "t@e.com", Password: "secret123",
		Type:        domain.CompanyCustomerType,
		Document:    "AB.CDE.FGH/0001-96", // wrong check digit
		CompanyName: "Test",
		Phone:       "11999999999",
	}

	_, err := domain.CreateCustomerToDomain(body)
	if err == nil {
		t.Fatal("expected error for invalid alphanumeric CNPJ check digit")
	}
}

func TestCreateCustomerToDomain_CNPJ_WrongLength(t *testing.T) {
	cases := []string{"1234567890123", "123456789012345"} // 13 and 15 chars

	for _, doc := range cases {
		body := dto.CreateCustomerRequest{
			Name: "Test", Email: "t@e.com", Password: "secret123",
			Type: domain.CompanyCustomerType, Document: doc,
			CompanyName: "Test", Phone: "11999999999",
		}
		_, err := domain.CreateCustomerToDomain(body)
		if err == nil {
			t.Errorf("expected error for CNPJ with wrong length %q", doc)
		}
	}
}

// ── Business rules ────────────────────────────────────────────────────────────

func TestCreateCustomerToDomain_COMPANY_MissingCompanyName(t *testing.T) {
	body := dto.CreateCustomerRequest{
		Name:     "Test",
		Email:    "t@e.com",
		Password: "secret123",
		Type:     domain.CompanyCustomerType,
		Document: "11.222.333/0001-81",
		Phone:    "11999999999",
		// CompanyName intentionally absent
	}

	_, err := domain.CreateCustomerToDomain(body)
	if err == nil {
		t.Fatal("expected error when company_name is missing for COMPANY type")
	}

	var badReq domain.BadRequestError
	if !asBadRequest(err, &badReq) {
		t.Errorf("expected BadRequestError, got %T", err)
	}
	if !strings.Contains(badReq.Message, "company_name") {
		t.Errorf("expected error message to mention 'company_name', got: %s", badReq.Message)
	}
}

func TestCreateCustomerToDomain_COMPANY_BlankCompanyName(t *testing.T) {
	body := dto.CreateCustomerRequest{
		Name: "Test", Email: "t@e.com", Password: "secret123",
		Type: domain.CompanyCustomerType, Document: "11.222.333/0001-81",
		CompanyName: "   ", // whitespace only
		Phone:       "11999999999",
	}

	_, err := domain.CreateCustomerToDomain(body)
	if err == nil {
		t.Fatal("expected error when company_name is blank")
	}
}

func TestCreateCustomerToDomain_SetsCorrectIDs(t *testing.T) {
	body := dto.CreateCustomerRequest{
		Name:     "Test",
		Email:    "t@e.com",
		Password: "secret123",
		Type:     domain.IndividualCustomerType,
		Document: "11144477735",
		Phone:    "11999999999",
	}

	c, _ := domain.CreateCustomerToDomain(body)

	if c.ID == "" || c.UserID == "" {
		t.Error("expected non-empty ID and UserID")
	}
	if c.ID != c.UserID {
		t.Error("expected Customer.ID == Customer.UserID")
	}
	if c.User.ID != c.ID {
		t.Error("expected User.ID == Customer.ID")
	}
}

func TestCreateCustomerToDomain_InvalidType(t *testing.T) {
	body := dto.CreateCustomerRequest{
		Name: "Test", Email: "t@e.com", Password: "secret123",
		Type: "UNKNOWN", Document: "11144477735", Phone: "11999999999",
	}

	_, err := domain.CreateCustomerToDomain(body)
	if err == nil {
		t.Fatal("expected error for invalid customer type")
	}
}

// ── helper ────────────────────────────────────────────────────────────────────

func asBadRequest(err error, target *domain.BadRequestError) bool {
	if br, ok := err.(domain.BadRequestError); ok {
		*target = br
		return true
	}
	return false
}
