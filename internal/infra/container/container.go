package container

import (
	"github.com/gin-gonic/gin"
)

type PingHttpHandler interface {
	Ping() gin.HandlerFunc
}

type UserHttpHandler interface {
	Create() gin.HandlerFunc
}

type WorkHttpHandler interface {
	Create() gin.HandlerFunc
}

type HTTP struct {
	PingHandler PingHttpHandler
	UserHandler UserHttpHandler
	WorkHandler WorkHttpHandler
}
