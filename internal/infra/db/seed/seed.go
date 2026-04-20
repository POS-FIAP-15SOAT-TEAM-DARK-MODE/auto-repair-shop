package seed

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/env"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const checkSeededQuery = `SELECT COUNT(*) FROM "work"`

// Run executes the database seed only if no seed data is present yet.
// Idempotency is guaranteed by checking for existing rows in the work table,
// which is exclusively populated by the seed (not by migrations).
func Run(ctx context.Context, db *sql.DB) {
	log := logger.Global()

	seeded, err := isAlreadySeeded(ctx, db)
	if err != nil {
		log.Error(fmt.Errorf("seed: failed to check seed state: %w", err))
		return
	}
	if seeded {
		log.Info("seed: data already present, skipping")
		return
	}

	log.Info("seed: starting database seeding")

	if err := runInTransaction(ctx, db, insertSeedData); err != nil {
		log.Error(fmt.Errorf("seed: failed to insert seed data: %w", err))
		return
	}

	log.Info("seed: database seeded successfully")
}

func isAlreadySeeded(ctx context.Context, db *sql.DB) (bool, error) {
	var count int
	if err := db.QueryRowContext(ctx, checkSeededQuery).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func runInTransaction(ctx context.Context, db *sql.DB, fn func(context.Context, *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err := fn(ctx, tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func insertSeedData(ctx context.Context, tx *sql.Tx) error {
	log := logger.Global()
	bcryptCost := env.GetInt("BCRYPT_COST", 12)

	if err := insertWorks(ctx, tx); err != nil {
		return err
	}
	log.Info("seed: works inserted")

	if err := insertSupplies(ctx, tx); err != nil {
		return err
	}
	log.Info("seed: supplies inserted")

	attendantID := uuid.NewString()
	if err := insertStaffUser(ctx, tx, attendantID, "attendant", "attendant@autorepairshop.com", "attendant123", "ATTENDANT", bcryptCost); err != nil {
		return err
	}
	log.Info("seed: attendant user inserted", zap.String("user_id", attendantID))

	mechanicID := uuid.NewString()
	if err := insertStaffUser(ctx, tx, mechanicID, "mechanic", "mechanic@autorepairshop.com", "mechanic123", "MECHANIC", bcryptCost); err != nil {
		return err
	}
	log.Info("seed: mechanic user inserted", zap.String("user_id", mechanicID))

	customerAUserID := uuid.NewString()
	customerAID := uuid.NewString()
	customerACPF := "529.982.247-25"
	if err := insertIndividualCustomer(ctx, tx, customerAUserID, customerAID, "João Silva", "joao.silva@email.com", customerACPF, "11999990001", bcryptCost); err != nil {
		return err
	}
	log.Info("seed: individual customer inserted", zap.String("customer_id", customerAID))

	vehicleAID := uuid.NewString()
	if err := insertVehicle(ctx, tx, vehicleAID, "ABC-1234", "Toyota", "Corolla", 2020, customerAID); err != nil {
		return err
	}
	log.Info("seed: vehicle inserted", zap.String("vehicle_id", vehicleAID))

	customerBUserID := uuid.NewString()
	customerBID := uuid.NewString()
	customerBCNPJ := "11.222.333/0001-81"
	if err := insertCompanyCustomer(ctx, tx, customerBUserID, customerBID, "Transportes Silva Ltda", "contato@transportessilva.com.br", customerBCNPJ, "Auto Transportes Silva", "11988880002", bcryptCost); err != nil {
		return err
	}
	log.Info("seed: company customer inserted", zap.String("customer_id", customerBID))

	vehicleBID := uuid.NewString()
	if err := insertVehicle(ctx, tx, vehicleBID, "XYZ1A23", "Ford", "Transit", 2022, customerBID); err != nil {
		return err
	}
	log.Info("seed: vehicle inserted", zap.String("vehicle_id", vehicleBID))

	return nil
}

// --- Works (billable services) ---

type workRow struct {
	name        string
	description string
	unitPrice   string
}

func insertWorks(ctx context.Context, tx *sql.Tx) error {
	works := []workRow{
		{"Troca de Óleo", "Troca de óleo do motor com filtro incluído", "120.00"},
		{"Rodízio de Pneus", "Rodízio completo dos quatro pneus", "80.00"},
		{"Alinhamento e Balanceamento", "Alinhamento de direção e balanceamento de rodas", "150.00"},
		{"Revisão de Freios", "Inspeção e ajuste do sistema de freios", "200.00"},
		{"Diagnóstico Eletrônico", "Leitura de falhas via scanner OBD-II", "100.00"},
	}

	const q = `INSERT INTO "work" (id, name, description, unit_price) VALUES ($1, $2, $3, $4)`

	for _, w := range works {
		if _, err := tx.ExecContext(ctx, q, uuid.NewString(), w.name, w.description, w.unitPrice); err != nil {
			return fmt.Errorf("insert work %q: %w", w.name, err)
		}
	}
	return nil
}

// --- Supplies (parts/stock) ---

type supplyRow struct {
	name        string
	description string
	unitPrice   string
	quantity    int
}

func insertSupplies(ctx context.Context, tx *sql.Tx) error {
	supplies := []supplyRow{
		{"Filtro de Óleo", "Filtro de óleo para motores 1.0 a 2.0", "35.00", 50},
		{"Pastilha de Freio Dianteira", "Jogo de pastilhas de freio dianteiras", "120.00", 30},
		{"Filtro de Ar", "Filtro de ar do motor", "45.00", 40},
		{"Fluido de Freio DOT 4", "Fluido de freio DOT 4 - 500ml", "28.00", 60},
		{"Correia Dentada", "Kit correia dentada com tensor e rolamento", "380.00", 15},
	}

	const q = `INSERT INTO supply (id, name, description, unit_price, stock_quantity) VALUES ($1, $2, $3, $4, $5)`

	for _, s := range supplies {
		if _, err := tx.ExecContext(ctx, q, uuid.NewString(), s.name, s.description, s.unitPrice, s.quantity); err != nil {
			return fmt.Errorf("insert supply %q: %w", s.name, err)
		}
	}
	return nil
}

// --- Staff users (ATTENDANT / MECHANIC) ---

func insertStaffUser(ctx context.Context, tx *sql.Tx, userID, name, email, rawPassword, roleName string, bcryptCost int) error {
	hash, err := hashPassword(rawPassword, bcryptCost)
	if err != nil {
		return fmt.Errorf("hash password for %q: %w", email, err)
	}

	const insertUser = `INSERT INTO "user" (id, name, email, password_hash) VALUES ($1, $2, $3, $4)`
	if _, err := tx.ExecContext(ctx, insertUser, userID, name, email, hash); err != nil {
		return fmt.Errorf("insert staff user %q: %w", email, err)
	}

	const assignRole = `INSERT INTO user_role (id, user_id, role_id) SELECT $1, $2, r.id FROM "role" r WHERE r.name = $3`
	if _, err := tx.ExecContext(ctx, assignRole, uuid.NewString(), userID, roleName); err != nil {
		return fmt.Errorf("assign role %q to user %q: %w", roleName, email, err)
	}

	return nil
}

// --- Individual customer (CPF) ---

func insertIndividualCustomer(ctx context.Context, tx *sql.Tx, userID, customerID, name, email, cpf, phone string, bcryptCost int) error {
	hash, err := hashPassword(cpf, bcryptCost)
	if err != nil {
		return fmt.Errorf("hash password for customer %q: %w", email, err)
	}

	const insertUser = `INSERT INTO "user" (id, name, email, password_hash) VALUES ($1, $2, $3, $4)`
	if _, err := tx.ExecContext(ctx, insertUser, userID, name, email, hash); err != nil {
		return fmt.Errorf("insert user for individual customer %q: %w", email, err)
	}

	const assignRole = `INSERT INTO user_role (id, user_id, role_id) SELECT $1, $2, r.id FROM "role" r WHERE r.name = $3`
	if _, err := tx.ExecContext(ctx, assignRole, uuid.NewString(), userID, "CUSTOMER"); err != nil {
		return fmt.Errorf("assign CUSTOMER role to %q: %w", email, err)
	}

	const insertCustomer = `INSERT INTO customer (id, user_id, type, cpf, phone) VALUES ($1, $2, 'INDIVIDUAL', $3, $4)`
	if _, err := tx.ExecContext(ctx, insertCustomer, customerID, userID, cpf, phone); err != nil {
		return fmt.Errorf("insert individual customer %q: %w", email, err)
	}

	return nil
}

// --- Company customer (CNPJ) ---

func insertCompanyCustomer(ctx context.Context, tx *sql.Tx, userID, customerID, name, email, cnpj, companyName, phone string, bcryptCost int) error {
	hash, err := hashPassword(cnpj, bcryptCost)
	if err != nil {
		return fmt.Errorf("hash password for company customer %q: %w", email, err)
	}

	const insertUser = `INSERT INTO "user" (id, name, email, password_hash) VALUES ($1, $2, $3, $4)`
	if _, err := tx.ExecContext(ctx, insertUser, userID, name, email, hash); err != nil {
		return fmt.Errorf("insert user for company customer %q: %w", email, err)
	}

	const assignRole = `INSERT INTO user_role (id, user_id, role_id) SELECT $1, $2, r.id FROM "role" r WHERE r.name = $3`
	if _, err := tx.ExecContext(ctx, assignRole, uuid.NewString(), userID, "CUSTOMER"); err != nil {
		return fmt.Errorf("assign CUSTOMER role to %q: %w", email, err)
	}

	const insertCustomer = `INSERT INTO customer (id, user_id, type, cnpj, company_name, phone) VALUES ($1, $2, 'COMPANY', $3, $4, $5)`
	if _, err := tx.ExecContext(ctx, insertCustomer, customerID, userID, cnpj, companyName, phone); err != nil {
		return fmt.Errorf("insert company customer %q: %w", email, err)
	}

	return nil
}

// --- Vehicle ---

func insertVehicle(ctx context.Context, tx *sql.Tx, vehicleID, plate, brand, model string, year int, customerID string) error {
	const q = `INSERT INTO vehicle (id, license_plate, brand, model, year, customer_id) VALUES ($1, $2, $3, $4, $5, $6)`
	if _, err := tx.ExecContext(ctx, q, vehicleID, plate, brand, model, year, customerID); err != nil {
		return fmt.Errorf("insert vehicle %q: %w", plate, err)
	}
	return nil
}

// --- Helpers ---

func hashPassword(raw string, cost int) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
