package user

import "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"

type userResponseDTO struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func mapUserToResponseDTO(user *domain.User) userResponseDTO {
	return userResponseDTO{
		Id:    user.Id,
		Name:  user.Name,
		Email: user.Email,
	}
}
