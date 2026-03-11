package customer

import "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"

type (
	service struct {
		customerRepository domain.CustomerRepository
	}
)

func Service(customerRepo domain.CustomerService) *service {
	return &service{
		customerRepository: customerRepo,
	}
}
