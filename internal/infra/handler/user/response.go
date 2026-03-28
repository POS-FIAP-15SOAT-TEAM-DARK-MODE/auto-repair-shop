package user

import "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"

type (
	userResponseDTO struct {
		Id    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	loggedUserResponseDTO struct {
		Token     string `json:"token"`
		ExpiresIn int    `json:"expires_in"`
	}
)

func mapUserToResponseDTO(user *domain.User) userResponseDTO {
	return userResponseDTO{
		Id:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}

func loggerUserToResponseDTO(loggedUser *domain.LoggedUser) *loggedUserResponseDTO {
	return &loggedUserResponseDTO{
		Token:     loggedUser.SessionToken,
		ExpiresIn: loggedUser.SessionExpiresIn,
	}
}
