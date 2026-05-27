//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/pkg/id"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/pkg/uow"
	supply "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/supply"
	supplyDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/supply/adapters"
	supplyInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/supply/interfaces"
	supplyRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/supply/repository"
	vehicle "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/vehicle"
	vehicleDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/vehicle/domain"
	vehicleInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/vehicle/interfaces"
	vehicleRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/vehicle/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	customerHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/customer"
	pingHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/ping"
	soHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/service_order"
	soHistoryHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/service_order_history"
	userHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/user"
	workHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/work"
	container "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http/middleware"
	customerRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/customer"
	soRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/service_order"
	soHistoryRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/service_order_history"
	userRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/user"
	workRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/work"
	workSOHistoryRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/work_service_order_history"
	uowV1 "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/routing"
	customerSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/customer"
	soSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/service_order"
	soHistorySvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/service_order_history"
	userSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/user"
	workSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/work"
)

var (
	testServer     *httptest.Server
	memUserRepo    domain.UserRepository
	memSupplyRepo  supplyInterfaces.SupplyService
	seedCustomerID string
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	ctx := context.Background()

	userRepository := userRepo.MemoryRepository()
	customerRepository := customerRepo.MemoryRepository()
	workRepository := workRepo.MemoryRepository()
	workSOHistoryRepository := workSOHistoryRepo.MemoryRepository()
	soHistoryStore := soHistoryRepo.NewMemoryStore()
	soRepository := soRepo.MemoryRepositoryWithHistory(soHistoryStore.Record)

	// MIGRATED
	vehicleRepository := vehicleRepo.NewInMemory()
	supplyRepository := supplyRepo.NewInMemory()
	// MIGRATED

	memUserRepo = userRepository

	noopUoW := &uowV1.UnitOfWork{}
	noopUoWV2 := uow.NewInMemoryUoW()
	expiresIn := 24 * time.Hour
	idGen := id.NewIDGenerator()

	// MIGRATED
	vehicleService := vehicle.NewService(noopUoWV2, vehicleRepository, idGen)
	supplyService := supply.NewService(noopUoWV2, supplyRepository, idGen)
	// MIGRATED

	userService := userSvc.Service(noopUoW, userRepository, expiresIn)
	customerService := customerSvc.Service(noopUoW, userRepository, customerRepository)
	workService := workSvc.Service(noopUoW, workRepository)
	soHistoryService := soHistorySvc.Service(noopUoW, soHistoryStore)
	soService := soSvc.Service(noopUoW, soRepository, workRepository, supplyService, soHistoryStore, workSOHistoryRepository, customerService, vehicleService)

	handlers := &container.HandlersWrapper{
		PingHandler:                pingHandler.HttpHandler(),
		UserHandler:                userHandler.HttpHandler(userService),
		CustomerHandler:            customerHandler.NewHandler(customerService),
		WorkHandler:                workHandler.HttpHandler(workService),
		ServiceOrderHandler:        soHandler.HttpHandler(soService),
		ServiceOrderHistoryHandler: soHistoryHandler.HttpHandler(soHistoryService),
		// MIGRATED
		VehicleHandler: vehicle.NewController(vehicleService),
		SupplyHandler:  supply.NewController(supplyService),
	}

	middlewares := &container.Middlewares{
		"Logger":       middleware.Logger(),
		"Recovery":     middleware.Recovery(),
		"ErrorHandler": middleware.ErrorHandler(),
	}

	memSupplyRepo = supplyService
	seedCustomerID = seedMemory(ctx, userRepository, customerRepository, vehicleRepository, workRepository, supplyService)
	router := routing.SetupRouter(handlers, middlewares)
	testServer = httptest.NewServer(router)
	defer testServer.Close()

	os.Exit(m.Run())
}

func hashPW(raw string) string {
	h, err := bcrypt.GenerateFromPassword([]byte(raw), 4)
	if err != nil {
		panic(fmt.Sprintf("seedMemory: bcrypt failed for %q: %v", raw, err))
	}
	return string(h)
}

func seedMemory(
	ctx context.Context,
	userRepository domain.UserRepository,
	customerRepository domain.CustomerRepository,
	vehicleRepository vehicleInterfaces.VehicleRepository,
	workRepository domain.WorkRepository,
	supplyRepository supplyInterfaces.SupplyService,
) string {
	// Staff users
	attendantID := uuid.NewString()
	_ = userRepository.Create(ctx, &domain.User{ID: attendantID, Name: "attendant", Email: "attendant@autorepairshop.com", Password: hashPW("attendant123")})
	_ = userRepository.AssignRole(ctx, attendantID, domain.ATTENDANT)

	mechanicID := uuid.NewString()
	_ = userRepository.Create(ctx, &domain.User{ID: mechanicID, Name: "mechanic", Email: "mechanic@autorepairshop.com", Password: hashPW("mechanic123")})
	_ = userRepository.AssignRole(ctx, mechanicID, domain.MECHANIC)

	// Customer João Silva (CPF sanitized = digits only)
	joaoUserID := uuid.NewString()
	joaoUser := &domain.User{ID: joaoUserID, Name: "João Silva", Email: "joao.silva@email.com", Password: hashPW("529.982.247-25")}
	_ = userRepository.Create(ctx, joaoUser)
	_ = userRepository.AssignRole(ctx, joaoUserID, domain.CUSTOMER)

	joaoCustomerID := uuid.NewString()
	_ = customerRepository.Create(ctx, &domain.Customer{
		ID:     joaoCustomerID,
		UserID: joaoUserID,
		Type:   domain.IndividualCustomerType,
		CPF:    "52998224725",
		Phone:  "11999990001",
		User:   joaoUser,
	})

	// Vehicle ABC-1234 (plate stored normalized, no hyphen)
	_ = vehicleRepository.Save(ctx, &vehicleDomain.Vehicle{
		ID:           uuid.NewString(),
		LicensePlate: "ABC1234",
		Brand:        "Toyota",
		Model:        "Corolla",
		Year:         2020,
		CustomerId:   joaoCustomerID,
	})

	// 5 seed works
	for _, w := range []struct{ name, desc, price string }{
		{"Troca de Óleo", "Troca de óleo do motor com filtro incluído", "120.00"},
		{"Rodízio de Pneus", "Rodízio completo dos quatro pneus do veículo", "80.00"},
		{"Alinhamento e Balanceamento", "Alinhamento de direção e balanceamento de rodas", "150.00"},
		{"Revisão de Freios", "Inspeção e ajuste do sistema de freios completo", "200.00"},
		{"Diagnóstico Eletrônico", "Leitura de falhas via scanner OBD-II do veículo", "100.00"},
	} {
		p, _ := decimal.NewFromString(w.price)
		_ = workRepository.Save(ctx, &domain.Work{ID: uuid.NewString(), Name: w.name, Description: w.desc, Price: p, Status: domain.ACTIVE})
	}

	// 5 seed supplies
	for _, s := range []struct {
		name, desc, price string
		qty               int
	}{
		{"Filtro de Óleo", "Filtro de óleo para motores 1.0 a 2.0 aspirados", "35.00", 50},
		{"Pastilha de Freio Dianteira", "Jogo de pastilhas de freio dianteiras completo", "120.00", 30},
		{"Filtro de Ar", "Filtro de ar do motor para veículos populares", "45.00", 40},
		{"Fluido de Freio DOT 4", "Fluido de freio DOT 4 - frasco de 500ml", "28.00", 60},
		{"Correia Dentada", "Kit correia dentada com tensor e rolamento incluso", "380.00", 15},
	} {
		p, _ := decimal.NewFromString(s.price)
		_, _ = supplyRepository.Create(ctx, supplyDomain.CreateSupply{Name: s.name, Description: s.desc, UnitPrice: p, StockQuantity: s.qty})
	}

	return joaoCustomerID
}
