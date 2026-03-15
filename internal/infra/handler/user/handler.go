package user

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/user/dto"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
)

type Handler struct {
	service domain.UserService
}

func NewHandler(service domain.UserService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateUser(c *gin.Context) {
	// TODO: consider remove the dto folder and put the files in the same package-level, avoiding this public methods
	req, err := dto.MapBodyToUserRequestDTO(c)
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
