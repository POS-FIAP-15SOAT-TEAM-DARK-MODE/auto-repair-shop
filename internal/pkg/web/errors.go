package web

import (
	"errors"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/gin-gonic/gin"
)

func Error(c *gin.Context, err error) {
	var badRequestError domain.BadRequestError
	var conflictError domain.ConflictError

	switch {
	case errors.As(err, &badRequestError):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": badRequestError.Message})
		return

	case errors.As(err, &conflictError):
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": conflictError.Message})
		return

	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	}

}
