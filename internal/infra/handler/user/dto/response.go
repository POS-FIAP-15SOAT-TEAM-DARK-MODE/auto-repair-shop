package dto

import "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"

type UserResponseDTO struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func MapUserToResponseDTO(user *domain.User) UserResponseDTO {
	return UserResponseDTO{
		Id:    user.Id,
		Name:  user.Name,
		Email: user.Email,
	}
}
