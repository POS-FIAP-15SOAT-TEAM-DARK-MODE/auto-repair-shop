package user

import (
	"github.com/gin-gonic/gin"

	services "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/user"
)

type Handler struct {
	Service services.UserService
}

func NewHandler(service services.UserService) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req services.CreateUserRequest

	// ShouldBindJSON é um método do Gin que tenta bindar o JSON da requisição para a struct fornecida.
	// Se o JSON for inválido ou não corresponder à struct, ele retorna um erro.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	userId, err := h.Service.CreateUser(req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"id": userId})
}
