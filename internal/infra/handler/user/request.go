package user

import (
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/json"
	"github.com/gin-gonic/gin"
)

type (
	userRequestDTO struct {
		Name            string `json:"name"`
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}

	loginRequestDTO struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	loggedUserResponseDTO struct {
		Token     string `json:"token"`
		ExpiresIn int    `json:"expires_in"`
	}
)

func mapBodyToUserRequestDTO(body *gin.Context) (*userRequestDTO, error) {
	var req userRequestDTO
	if err := body.ShouldBindJSON(&req); err != nil {
		return nil, json.CheckJsonError(err)
	}
	return &req, nil
}

func mapUserRequestDTOToDomain(req *userRequestDTO) *domain.User {
	return domain.NewUser(req.Name, req.Email, req.Password)
}

func (u *userRequestDTO) Validate() error {
	if strings.TrimSpace(u.Password) != strings.TrimSpace(u.ConfirmPassword) {
		return domain.ErrUserPasswordDontMatch
	}
	return nil
}

func mapBodyToLoginRequestDTO(body *gin.Context) (*loginRequestDTO, error) {
	var req loginRequestDTO
	if err := body.ShouldBindJSON(&req); err != nil {
		return nil, json.CheckJsonError(err)
	}
	return &req, nil
}

func (l *loginRequestDTO) toLoggedUserDomain() *domain.LoggedUser {
	return &domain.LoggedUser{
		User: domain.User{
			Email:    l.Email,
			Password: l.Password,
		},
	}
}

func loggerUserToResponseDTO(loggedUser *domain.LoggedUser) *loggedUserResponseDTO {
	return &loggedUserResponseDTO{
		Token:     loggedUser.SessionToken,
		ExpiresIn: loggedUser.SessionExpiresIn,
	}
}
