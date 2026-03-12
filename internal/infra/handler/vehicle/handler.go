package vehicle

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Create(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
