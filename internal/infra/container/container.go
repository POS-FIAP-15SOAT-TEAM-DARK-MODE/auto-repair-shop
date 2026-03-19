package container

import (
	"github.com/gin-gonic/gin"
)

type (
	PingHttpHandler interface {
		Ping() gin.HandlerFunc
	}

	UserHttpHandler interface {
		Create() gin.HandlerFunc
	}

	CustomerHttpHandler interface {
		Create() gin.HandlerFunc
	}

	WorkHttpHandler interface {
		Create() gin.HandlerFunc
		List() gin.HandlerFunc
	}

	VehicleHttpHandler interface {
		Create() gin.HandlerFunc
	}
)

type HTTP struct {
	PingHandler     PingHttpHandler
	UserHandler     UserHttpHandler
	CustomerHandler CustomerHttpHandler
	WorkHandler     WorkHttpHandler
	VehicleHandler  VehicleHttpHandler
}
