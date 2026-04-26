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
//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceOrderHistoryHandler --with-expecter
type (
	PingHandler interface {
		Ping(*gin.Context)
	}

	UserHandler interface {
		Create(*gin.Context)
		Login(*gin.Context)
		UpdateRole(*gin.Context)
	}

	CustomerHandler interface {
		Create(*gin.Context)
		GetByID(*gin.Context)
		GetByDocument(*gin.Context)
		Update(*gin.Context)
		Delete(*gin.Context)
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
		FindByCustomer(*gin.Context)
		Update(*gin.Context)
		Delete(*gin.Context)
	}

	SupplyHandler interface {
		Create(*gin.Context)
		List(*gin.Context)
		Update(*gin.Context)
		Delete(*gin.Context)
	}

	ServiceOrderHandler interface {
		Create(*gin.Context)
		List(*gin.Context)
		GetWorks(*gin.Context)
		AddWork(*gin.Context)
		DeleteWork(*gin.Context)
		GetSupplies(*gin.Context)
		AddSupplies(*gin.Context)
		DeleteSupply(*gin.Context)
		SendToCustomerApproval(*gin.Context)
		SendToDiagnosis(*gin.Context)
		Accept(*gin.Context)
		Reject(*gin.Context)
		Deliver(*gin.Context)
		Cancel(*gin.Context)
		GetFullByID(*gin.Context)
		GetAverageExecutionTime(*gin.Context)
		GetStatus(*gin.Context)
	}

	ServiceOrderHistoryHandler interface {
		GetHistoryByID(*gin.Context)
	}
)

type HandlersWrapper struct {
	PingHandler                PingHandler
	UserHandler                UserHandler
	CustomerHandler            CustomerHandler
	WorkHandler                WorkHandler
	VehicleHandler             VehicleHandler
	SupplyHandler              SupplyHandler
	ServiceOrderHandler        ServiceOrderHandler
	ServiceOrderHistoryHandler ServiceOrderHistoryHandler
}
