package dto

import (
	"fmt"
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/gin-gonic/gin"
)

type UserRequestDTO struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

func MapBodyToUserRequestDTO(body *gin.Context) (*UserRequestDTO, error) {
	var req UserRequestDTO
	if err := body.ShouldBindJSON(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

func MapUserRequestDTOToDomain(req *UserRequestDTO) *domain.User {
	return domain.NewUser(req.Name, req.Email, req.Password)
}

func (u *UserRequestDTO) Validate() error {
	if strings.TrimSpace(u.Password) != strings.TrimSpace(u.ConfirmPassword) {
		return fmt.Errorf("passwords do not match")
	}
	return nil
}
