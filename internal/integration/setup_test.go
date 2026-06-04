//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"

	authPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/auth"
	authDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/auth/domain"
	authInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/auth/interfaces"
	authRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/auth/repository"
	customerPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/customer"
	customerAdapters "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/customer/adapters"
	customerDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/customer/domain"
	customerInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/customer/interfaces"
	customerRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/customer/repository"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/id"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	serviceOrderPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order"
	soRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/repository"
	soHistoryPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order_history"
	soHistoryRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order_history/repository"
	supplyPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/supply"
	supplyAdapters "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/supply/adapters"
	supplyInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/supply/interfaces"
	supplyRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/supply/repository"
	vehiclePkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/vehicle"
	vehicleDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/vehicle/domain"
	vehicleInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/vehicle/interfaces"
	vehicleRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/vehicle/repository"
	workPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/work"
	workDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/work/domain"
	workInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/work/interfaces"
	workRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/work/repository"

	pingHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/ping"
	container "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http/middleware"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/routing"
)

var (
	testServer     *httptest.Server
	memUserRepo    authInterfaces.AuthRepository
	memSupplyRepo  supplyInterfaces.SupplyService
	seedCustomerID string
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	ctx := context.Background()
	idGen := id.NewIDGenerator()
	noopUoW := uow.NewInMemoryUoW()

	// Auth
	userRepository := authRepo.NewInMemory()
	authSvc := authPkg.NewService(noopUoW, userRepository, idGen)
	authCtrl := authPkg.NewController(authSvc)

	// Customer
	custRepository := customerRepo.NewInMemory()
	custSvc := customerPkg.NewService(noopUoW, custRepository, userRepository)
	custCtrl := customerPkg.NewController(custSvc)

	// Work
	wRepository := workRepo.NewInMemory()
	workSvc := workPkg.NewService(noopUoW, wRepository)
	workCtrl := workPkg.NewController(workSvc)

	// Vehicle
	vehicleRepository := vehicleRepo.NewInMemory()
	vehicleSvc := vehiclePkg.NewService(noopUoW, vehicleRepository, idGen)
	vehicleCtrl := vehiclePkg.NewController(vehicleSvc)

	// Supply
	supplyRepository := supplyRepo.NewInMemory()
	supplySvc := supplyPkg.NewService(noopUoW, supplyRepository, idGen)
	supplyCtrl := supplyPkg.NewController(supplySvc)

	// Service order history
	soHistoryStore := soHistoryRepo.NewMemoryStore()
	soHistorySvc := soHistoryPkg.NewService(soHistoryStore)
	soHistoryCtrl := soHistoryPkg.NewController(soHistorySvc)

	// Service order
	soRepository := soRepo.NewSOMemoryWithHistory(soHistoryStore.Record)
	wsoHistoryRepository := soRepo.NewWSOHistoryMemory()
	customerFinder := newCustomerFinder(custRepository)
	var wRepo workInterfaces.WorkRepository = wRepository
	soSvc := serviceOrderPkg.NewService(noopUoW, soRepository, wRepo, supplySvc, wsoHistoryRepository, customerFinder, vehicleSvc)
	soCtrl := serviceOrderPkg.NewController(soSvc)

	memUserRepo = userRepository
	memSupplyRepo = supplySvc

	seedCustomerID = seedMemory(ctx, userRepository, custRepository, vehicleRepository, wRepository, supplySvc)

	handlers := &container.HandlersWrapper{
		PingHandler:                pingHandler.HttpHandler(),
		UserHandler:                authCtrl,
		CustomerHandler:            custCtrl,
		WorkHandler:                workCtrl,
		VehicleHandler:             vehicleCtrl,
		SupplyHandler:              supplyCtrl,
		ServiceOrderHandler:        soCtrl,
		ServiceOrderHistoryHandler: soHistoryCtrl,
	}

	middlewares := &container.Middlewares{
		"Logger":       middleware.Logger(),
		"Recovery":     middleware.Recovery(),
		"ErrorHandler": middleware.ErrorHandler(),
	}

	router := routing.SetupRouter(handlers, middlewares)
	testServer = httptest.NewServer(router)
	defer testServer.Close()

	os.Exit(m.Run())
}

type customerFinderAdapter struct {
	repo customerInterfaces.CustomerRepository
}

func newCustomerFinder(repo customerInterfaces.CustomerRepository) serviceOrderPkg.CustomerFinder {
	return &customerFinderAdapter{repo: repo}
}

func (f *customerFinderAdapter) GetByID(ctx context.Context, customerID string) (customerDomain.Customer, error) {
	items, err := f.repo.List(ctx, customerAdapters.ListCustomerParams{ID: customerID})
	if err != nil {
		return customerDomain.Customer{}, err
	}
	if len(items) == 0 {
		return customerDomain.Customer{}, customerDomain.ErrCustomerNotFound
	}
	return items[0], nil
}

func (f *customerFinderAdapter) GetByUserID(ctx context.Context, userID string) (customerDomain.Customer, error) {
	items, err := f.repo.List(ctx, customerAdapters.ListCustomerParams{UserID: userID})
	if err != nil {
		return customerDomain.Customer{}, err
	}
	if len(items) == 0 {
		return customerDomain.Customer{}, customerDomain.ErrCustomerNotFound
	}
	return items[0], nil
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
	userRepository authInterfaces.AuthRepository,
	custRepository customerInterfaces.CustomerRepository,
	vehicleRepository vehicleInterfaces.VehicleRepository,
	wRepository workInterfaces.WorkRepository,
	supplySvc supplyInterfaces.SupplyService,
) string {
	// Staff users
	attendantID := uuid.NewString()
	_ = userRepository.Save(ctx, &authDomain.User{
		ID: attendantID, Name: "attendant",
		CoreUser: authDomain.CoreUser{Email: "attendant@autorepairshop.com", Password: hashPW("attendant123")},
	})
	_ = userRepository.SaveUserRole(ctx, attendantID, authDomain.ATTENDANT)

	mechanicID := uuid.NewString()
	_ = userRepository.Save(ctx, &authDomain.User{
		ID: mechanicID, Name: "mechanic",
		CoreUser: authDomain.CoreUser{Email: "mechanic@autorepairshop.com", Password: hashPW("mechanic123")},
	})
	_ = userRepository.SaveUserRole(ctx, mechanicID, authDomain.MECHANIC)

	// Customer João Silva (CPF sanitized = digits only)
	joaoUserID := uuid.NewString()
	_ = userRepository.Save(ctx, &authDomain.User{
		ID: joaoUserID, Name: "João Silva",
		CoreUser: authDomain.CoreUser{Email: "joao.silva@email.com", Password: hashPW("529.982.247-25")},
	})
	_ = userRepository.SaveUserRole(ctx, joaoUserID, authDomain.CUSTOMER)

	joaoCustomerID := uuid.NewString()
	_ = custRepository.Save(ctx, &customerDomain.Customer{
		ID:     joaoCustomerID,
		UserID: joaoUserID,
		Type:   customerDomain.IndividualCustomerType,
		CPF:    "52998224725",
		Phone:  "11999990001",
	})

	// Vehicle ABC-1234
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
		_ = wRepository.Save(ctx, &workDomain.Work{ID: uuid.NewString(), Name: w.name, Description: w.desc, Price: p, Status: workDomain.ACTIVE})
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
		_, _ = supplySvc.Create(ctx, supplyAdapters.CreateSupply{Name: s.name, Description: s.desc, UnitPrice: p, StockQuantity: s.qty})
	}

	return joaoCustomerID
}
