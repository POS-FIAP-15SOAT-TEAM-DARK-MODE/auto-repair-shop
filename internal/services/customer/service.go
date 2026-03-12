package customer

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/customer/dto"
	"github.com/google/uuid"
)

type (
	service struct {
		customerRepository domain.CustomerRepository
		userRepository     domain.UserRepository
	}
)

func Service(customerRepo domain.CustomerRepository, userRepo domain.UserRepository) *service {
	return &service{
		customerRepository: customerRepo,
		userRepository:     userRepo,
	}
}

func (s *service) Create(body dto.CreateCustomerRequest) (domain.Customer, error) {
	customer := domain.CreateCustomerToDomain(body)

	id := uuid.New().String()
	customer.User.ID = id

	_, err := s.userRepository.Create(customer.User)
	if err != nil {
		//TODO handle error: if user already exists, return conflict error, otherwise return internal server error
		return domain.Customer{}, err
	}

	customer.ID = id
	customer.UserID = id

	if err := s.customerRepository.Create(&customer); err != nil {
		//TODO handle error: if customer already exists, return conflict error, otherwise return internal server error
		return domain.Customer{}, err
	}

	return domain.CreateCustomerToDomain(body), nil
}
