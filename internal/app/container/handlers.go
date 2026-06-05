package container

import (
	authInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/auth/interfaces"
	customerInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/customer/interfaces"
	soInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/interfaces"
	soHistoryInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order_history/interfaces"
	supplyInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/supply/interfaces"
	vehicleInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/vehicle/interfaces"
	workInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/work/interfaces"

	"github.com/gin-gonic/gin"
)

type HandlersWrapper struct {
	UserHandler                authInterfaces.AuthHTTPController
	CustomerHandler            customerInterfaces.CustomerHTTPController
	WorkHandler                workInterfaces.WorkHTTPController
	VehicleHandler             vehicleInterfaces.VehicleHTTPController
	SupplyHandler              supplyInterfaces.SupplyHTTPController
	ServiceOrderHandler        soInterfaces.ServiceOrderHTTPController
	ServiceOrderHistoryHandler soHistoryInterfaces.ServiceOrderHistoryHTTPController
}

type Middlewares map[string]gin.HandlerFunc
