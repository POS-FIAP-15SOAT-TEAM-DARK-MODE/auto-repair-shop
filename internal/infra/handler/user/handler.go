package user

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/bind"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/json"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
)

type (
	handler struct {
		service domain.UserService
	}

	loginRequestDTO struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
)

func HttpHandler(service domain.UserService) *handler {
	return &handler{service: service}
}

func (h *handler) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		req, err := bind.BindGenericJSON[loginRequestDTO](c)
		if err != nil {
			status, response := web.Error(json.CheckJsonError(err))
			logger.Of(ctx).Debug("Failed to bind login payload",
				zap.String("operation", "login"),
				zap.Error(err),
				zap.String("entity", "user"),
			)
			c.JSON(status, response)
			return
		}

		if err := req.validate(); err != nil {
			status, response := web.Error(err)
			logger.Of(ctx).Debug("Login validation failed",
				zap.String("operation", "login"),
				zap.Error(err),
				zap.String("entity", "user"),
			)
			c.JSON(status, response)
			return
		}

		user := req.dtoToDomain()

		response, err := h.service.Login(ctx, user)
		if err != nil {
			status, response := web.Error(err)
			logger.Of(ctx).Error(err)
			logger.Of(ctx).Debug("Login failed in service layer",
				zap.String("operation", "login"),
				zap.Error(err),
				zap.String("entity", "user"),
			)
			c.JSON(status, response)
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

func (h *handler) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		req, err := mapBodyToUserRequestDTO(c)
		if err != nil {
			status, response := web.Error(err)
			logger.Of(ctx).Debug("Failed to bind user creation payload",
				zap.String("operation", "create_user"),
				zap.Error(err),
				zap.String("entity", "user"),
			)
			c.JSON(status, response)
			return
		}

		if err := req.Validate(); err != nil {
			status, response := web.Error(err)
			logger.Of(ctx).Debug("User validation failed",
				zap.String("operation", "create_user"),
				zap.Error(err),
				zap.String("entity", "user"),
			)
			c.JSON(status, response)
			return
		}

		user := mapUserRequestDTOToDomain(req)
		if err = h.service.Create(ctx, user); err != nil {
			status, response := web.Error(err)
			logger.Of(ctx).Error(err)
			logger.Of(ctx).Debug("User creation failed in service layer",
				zap.String("operation", "create_user"),
				zap.Error(err),
				zap.String("entity", "user"),
			)
			c.JSON(status, response)
			return
		}

		logger.Of(ctx).Debug("create response", zap.Any("service", user))
		c.JSON(http.StatusCreated, mapUserToResponseDTO(user))
	}
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

func (l *loginRequestDTO) dtoToDomain() *domain.User {
	return &domain.User{
		Email:    l.Email,
		Password: l.Password,
	}
}
