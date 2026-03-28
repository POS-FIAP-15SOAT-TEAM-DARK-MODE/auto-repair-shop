package customer

import (
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	pkgjson "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/json"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
	"github.com/gin-gonic/gin"
)

type handler struct {
	service domain.CustomerService
}

func NewHandler(service domain.CustomerService) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var body createCustomerRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		status, response := web.Error(pkgjson.CheckJsonError(err))
		c.JSON(status, response)
		return
	}

	if err := body.validate(); err != nil {
		status, response := web.Error(err)
		c.JSON(status, response)
		return
	}

	customer, err := body.Domain()
	if err != nil {
		status, response := web.Error(err)
		c.JSON(status, response)
		return
	}

	if err = h.service.Create(ctx, customer); err != nil {
		status, response := web.Error(err)
		c.JSON(status, response)
		return
	}

	c.JSON(http.StatusCreated, toResponse(customer))
}

func (h *handler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()

	req := getCustomerByIDRequest{ID: c.Param("id")}
	if err := req.validate(); err != nil {
		status, response := web.Error(err)
		c.JSON(status, response)
		return
	}

	customer, err := h.service.GetByID(ctx, req.ID)
	if err != nil {
		status, response := web.Error(err)
		c.JSON(status, response)
		return
	}

	c.JSON(http.StatusOK, toResponse(customer))
}

func (h *handler) GetByDocument(c *gin.Context) {
	ctx := c.Request.Context()

	req := getCustomerByDocumentRequest{Document: c.Query("document")}
	if err := req.validate(); err != nil {
		status, response := web.Error(err)
		c.JSON(status, response)
		return
	}

	customer, err := h.service.GetByDocument(ctx, req.Document)
	if err != nil {
		status, response := web.Error(err)
		c.JSON(status, response)
		return
	}

	c.JSON(http.StatusOK, toResponse(customer))
}
