package vehicle

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type handler struct {
}

func HttpHandler() *handler {
	return &handler{}
}

func (h *handler) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Status(http.StatusCreated)
	}
}
