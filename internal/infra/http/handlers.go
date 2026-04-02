package http

import (
	"github.com/gin-gonic/gin"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=PingHandler --with-expecter
//go:generate go run github.com/vektra/mockery/v2@latest --name=UserHandler --with-expecter
//go:generate go run github.com/vektra/mockery/v2@latest --name=CustomerHandler --with-expecter
//go:generate go run github.com/vektra/mockery/v2@latest --name=WorkHandler --with-expecter
//go:generate go run github.com/vektra/mockery/v2@latest --name=VehicleHandler --with-expecter
//go:generate go run github.com/vektra/mockery/v2@latest --name=SupplyHandler --with-expecter
//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceOrderHandler --with-expecter
type (
	PingHandler interface {
		Ping(*gin.Context)
	}

	UserHandler interface {
		Create(*gin.Context)
		Login(*gin.Context)
	}

	CustomerHandler interface {
		Create(*gin.Context)
		GetByID(*gin.Context)
		GetByDocument(*gin.Context)
	}

	WorkHandler interface {
		Create(*gin.Context)
		List(*gin.Context)
		Update(*gin.Context)
		Delete(*gin.Context)
	}

	VehicleHandler interface {
		Create(*gin.Context)
		FindByLicensePlate(*gin.Context)
	}
	SupplyHandler interface {
		Create(*gin.Context)
	}

	ServiceOrderHandler interface {
		Create(*gin.Context)
	}
)

type HandlersWrapper struct {
	PingHandler         PingHandler
	UserHandler         UserHandler
	CustomerHandler     CustomerHandler
	WorkHandler         WorkHandler
	VehicleHandler      VehicleHandler
	SupplyHandler       SupplyHandler
	ServiceOrderHandler ServiceOrderHandler
}
