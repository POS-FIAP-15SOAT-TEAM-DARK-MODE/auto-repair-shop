package user

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/user/dto"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/bind"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/json"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
)

type (
	Handler struct {
		service domain.UserService
	}

	loginRequestDTO struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
)

func NewHandler(service domain.UserService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) LoginUser(c *gin.Context) {
	req, err := bind.BindGenericJSON[loginRequestDTO](c)
	if err != nil {
		status, response := web.Error(json.CheckJsonError(err))
		// TODO: add logging
		c.JSON(status, response)
		return
	}

	if err := req.validate(); err != nil {
		status, response := web.Error(err)
		// TODO: add logging
		c.JSON(status, response)
		return
	}

	user, err := req.dtoToDomain()
	if err != nil {
		status, response := web.Error(err)
		// TODO: add logging
		c.JSON(status, response)
		return
	}

	response, err := h.service.Login(user)
	if err != nil {
		status, response := web.Error(err)
		// TODO: add logging
		c.JSON(status, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) CreateUser(c *gin.Context) {
	// TODO: consider remove the dto folder and put the files in the same package-level, avoiding this public methods
	req, err := bind.BindGenericJSON[dto.UserRequestDTO](c)
	if err != nil {
		status, response := web.Error(err)
		// TODO: add logging
		c.JSON(status, response)
		return
	}

	if err := req.Validate(); err != nil {
		status, response := web.Error(err)
		// TODO: add logging
		c.JSON(status, response)
		return
	}

	user := dto.MapUserRequestDTOToDomain(req)
	user, err = h.service.Create(user)
	if err != nil {
		status, response := web.Error(err)
		// TODO: add logging
		c.JSON(status, response)
		return
	}

	c.JSON(http.StatusCreated, dto.MapUserToResponseDTO(user))
}

func (l *loginRequestDTO) validate() error {
	if strings.TrimSpace(l.Email) == "" {
		return domain.ErrEmptyEmail
	}

	if strings.TrimSpace(l.Password) == "" {
		return domain.ErrEmptyPassword
	}

	return nil
}

func (l *loginRequestDTO) dtoToDomain() (*domain.User, error) {
	user := &domain.User{
		Email:    l.Email,
		Password: l.Password,
	}

	if err := user.HashPassword(); err != nil {
		return nil, err
	}

	return user, nil
}
