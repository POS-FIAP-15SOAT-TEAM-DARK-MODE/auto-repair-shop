package routing

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/api"
)

func mountSwaggerUI(router *gin.Engine) {
	router.GET("/swagger.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", api.SwaggerBytes)
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger.yaml")))
}
