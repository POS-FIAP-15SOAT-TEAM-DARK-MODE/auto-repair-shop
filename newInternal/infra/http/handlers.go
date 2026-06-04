package http

import (
	authInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/auth/interfaces"
	customerInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/customer/interfaces"
	soInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order/interfaces"
	soHistoryInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order_history/interfaces"
	supplyInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/supply/interfaces"
	vehicleInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/vehicle/interfaces"
	workInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/work/interfaces"
	"github.com/gin-gonic/gin"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=PingHandler --with-expecter
type PingHandler interface {
	Ping(*gin.Context)
}

type HandlersWrapper struct {
	PingHandler                PingHandler
	UserHandler                authInterfaces.AuthHTTPController
	CustomerHandler            customerInterfaces.CustomerHTTPController
	WorkHandler                workInterfaces.WorkHTTPController
	VehicleHandler             vehicleInterfaces.VehicleHTTPController
	SupplyHandler              supplyInterfaces.SupplyHTTPController
	ServiceOrderHandler        soInterfaces.ServiceOrderHTTPController
	ServiceOrderHistoryHandler soHistoryInterfaces.ServiceOrderHistoryHTTPController
}
