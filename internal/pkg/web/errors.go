package web

import (
	"errors"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/gin-gonic/gin"
)

func Error(c *gin.Context, err error) {
	var badRequestError domain.ValidationError
	var conflictError domain.ConflictError
	var unprocessableError domain.BusinessRuleError

	switch {
	case errors.As(err, &badRequestError):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": badRequestError.Message})

	case errors.As(err, &conflictError):
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": conflictError.Message})

	case errors.As(err, &unprocessableError):
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": unprocessableError.Message})

	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	}
}
