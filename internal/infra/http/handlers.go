package http

import (
	"github.com/gin-gonic/gin"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=PingHandler --with-expecter
type PingHandler interface {
	Ping() gin.HandlerFunc
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=UserHandler --with-expecter
type UserHandler interface {
	Create() gin.HandlerFunc
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=CustomerHandler --with-expecter
type CustomerHandler interface {
	Create() gin.HandlerFunc
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=WorkHandler --with-expecter
type WorkHandler interface {
	Create() gin.HandlerFunc
	List() gin.HandlerFunc
	Update() gin.HandlerFunc
	Delete() gin.HandlerFunc
}

type HandlersWrapper struct {
	PingHandler     PingHandler
	UserHandler     UserHandler
	CustomerHandler CustomerHandler
	WorkHandler     WorkHandler
}
