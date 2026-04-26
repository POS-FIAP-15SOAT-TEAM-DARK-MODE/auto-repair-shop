package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/json"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
)

type handler struct {
	service domain.UserService
}

func HttpHandler(service domain.UserService) *handler {
	return &handler{service: service}
}

func (h *handler) Login(c *gin.Context) {
	ctx := c.Request.Context()
	req, err := mapBodyToLoginRequestDTO(c)
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

	user := req.toLoggedUserDomain()

	if err := h.service.Login(ctx, user); err != nil {
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

	response := loggerUserToResponseDTO(user)

	c.JSON(http.StatusOK, response)
}

func (h *handler) Create(c *gin.Context) {
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

	logger.Of(ctx).Debug("create response",
		zap.String("operation", "create_user"),
		zap.String("entity", "user"),
		zap.String("user_id", user.ID),
		zap.String("email", user.Email),
	)
	c.JSON(http.StatusCreated, mapUserToResponseDTO(user))
}

func (h *handler) UpdateRole(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	req, err := mapBodyToUpdateRoleRequestDTO(c)
	if err != nil {
		status, response := web.Error(json.CheckJsonError(err))
		logger.Of(ctx).Debug("Failed to bind user role payload",
			zap.String("operation", "update_user_role"),
			zap.Error(err),
			zap.String("entity", "user"),
		)
		c.JSON(status, response)
		return
	}

	if err = req.Validate(); err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Debug("User role validation failed",
			zap.String("operation", "update_user_role"),
			zap.Error(err),
			zap.String("entity", "user"),
		)
		c.JSON(status, response)
		return
	}

	if err = h.service.UpdateRole(ctx, id, req.Role); err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Error(err)
		logger.Of(ctx).Debug("User role update failed in service layer",
			zap.String("operation", "update_user_role"),
			zap.Error(err),
			zap.String("entity", "user"),
		)
		c.JSON(status, response)
		return
	}

	c.Status(http.StatusNoContent)
}
