package http

import (
	supplyInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/supply/interfaces"
	vehicleInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/vehicle/interfaces"
	"github.com/gin-gonic/gin"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=PingHandler --with-expecter
//go:generate go run github.com/vektra/mockery/v2@latest --name=UserHandler --with-expecter
//go:generate go run github.com/vektra/mockery/v2@latest --name=CustomerHandler --with-expecter
//go:generate go run github.com/vektra/mockery/v2@latest --name=WorkHandler --with-expecter
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

	ServiceOrderHandler interface {
		Create(*gin.Context)
		List(*gin.Context)
		GetWorks(*gin.Context)
		AddWork(*gin.Context)
		DeleteWork(*gin.Context)
		GetSupplies(*gin.Context)
		AddSupplies(*gin.Context)
		DeleteSupply(*gin.Context)
		Receive(*gin.Context)
		SendToCustomerApproval(*gin.Context)
		SendToDiagnosis(*gin.Context)
		Finish(*gin.Context)
		Accept(*gin.Context)
		Reject(*gin.Context)
		Deliver(*gin.Context)
		Cancel(*gin.Context)
		GetFullByID(*gin.Context)
		GetAverageExecutionTime(*gin.Context)
		NextWork(*gin.Context)
		CancelWork(*gin.Context)
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
	VehicleHandler             vehicleInterfaces.VehicleHTTPController
	SupplyHandler              supplyInterfaces.SupplyHTTPController
	ServiceOrderHandler        ServiceOrderHandler
	ServiceOrderHistoryHandler ServiceOrderHistoryHandler
}
