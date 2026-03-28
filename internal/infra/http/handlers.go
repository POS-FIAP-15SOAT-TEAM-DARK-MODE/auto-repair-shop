package http

import (
	"github.com/gin-gonic/gin"
)

type (
	//go:generate go run github.com/vektra/mockery/v2@latest --name=PingHandler --with-expecter
	PingHandler interface {
		Ping() gin.HandlerFunc
	}

	//go:generate go run github.com/vektra/mockery/v2@latest --name=UserHandler --with-expecter
	UserHandler interface {
		Create() gin.HandlerFunc
		Login() gin.HandlerFunc
	}

	//go:generate go run github.com/vektra/mockery/v2@latest --name=CustomerHandler --with-expecter
	CustomerHandler interface {
		Create() gin.HandlerFunc
	}

	//go:generate go run github.com/vektra/mockery/v2@latest --name=WorkHandler --with-expecter
	WorkHandler interface {
		Create() gin.HandlerFunc
		List() gin.HandlerFunc
		Update() gin.HandlerFunc
		Delete() gin.HandlerFunc
	}

	//go:generate go run github.com/vektra/mockery/v2@latest --name=VehicleHandler --with-expecter
	VehicleHandler interface {
		Create() gin.HandlerFunc
	}
)

type HandlersWrapper struct {
	PingHandler     PingHandler
	UserHandler     UserHandler
	CustomerHandler CustomerHandler
	WorkHandler     WorkHandler
	VehicleHandler  VehicleHandler
}
