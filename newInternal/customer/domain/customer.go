package domain

import (
	userDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/auth/domain"
)

type CustomerType string

const (
	IndividualCustomerType CustomerType = "INDIVIDUAL"
	CompanyCustomerType    CustomerType = "COMPANY"
)

func (ct CustomerType) String() string {
	return string(ct)
}

type Customer struct {
	ID          string
	UserID      string
	Type        CustomerType
	CPF         string
	CNPJ        string
	CompanyName string
	Phone       string
	User        *userDomain.User
}
