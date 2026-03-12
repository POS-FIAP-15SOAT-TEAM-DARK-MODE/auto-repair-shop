package user

import (
	"github.com/gin-gonic/gin"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/user/dto"
	services "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/user"
)

type Handler struct {
	Service services.UserService
}

func NewHandler(service services.UserService) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req dto.UserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	user, err := h.Service.CreateUser(req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"user": dto.UserResponse{Id: user.Id, Name: user.Name, Email: user.Email}})
}
