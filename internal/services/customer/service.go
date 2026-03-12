package customer

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/customer/dto"
)

type (
	service struct {
		customerRepository domain.CustomerRepository
	}
)

func Service(customerRepo domain.CustomerRepository) *service {
	return &service{
		customerRepository: customerRepo,
	}
}

func (s *service) Create(body dto.CreateCustomerRequest) (domain.Customer, error) {
	return domain.CreateCustomerToDomain(body), nil
}
