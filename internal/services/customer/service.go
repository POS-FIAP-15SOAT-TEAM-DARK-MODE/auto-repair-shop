package customer

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/customer/dto"
)

type service struct {
	transactor domain.Transactor
}

func Service(transactor domain.Transactor) *service {
	return &service{transactor: transactor}
}

func (s *service) Create(body dto.CreateCustomerRequest) (domain.Customer, error) {
	customer, err := domain.CreateCustomerToDomain(body)
	if err != nil {
		return domain.Customer{}, domain.BadRequestError{Message: err.Error()}
	}

	err = s.transactor.WithTransaction(context.Background(), func(userRepo domain.UserRepository, customerRepo domain.CustomerRepository) error {
		if _, err := userRepo.Create(customer.User); err != nil {
			return err
		}
		return customerRepo.Create(&customer)
	})
	if err != nil {
		return domain.Customer{}, err
	}

	return customer, nil
}
