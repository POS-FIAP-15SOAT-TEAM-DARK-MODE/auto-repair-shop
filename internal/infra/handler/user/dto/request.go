package dto

import (
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type (
	UserRequestDTO struct {
		Name            string `json:"name"`
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}
)

func MapUserRequestDTOToDomain(req *UserRequestDTO) *domain.User {
	return domain.NewUser(req.Name, req.Email, req.Password)
}

func (u *UserRequestDTO) Validate() error {
	if strings.TrimSpace(u.Password) != strings.TrimSpace(u.ConfirmPassword) {
		return domain.ErrPasswordDontMatch
	}
	return nil
}
