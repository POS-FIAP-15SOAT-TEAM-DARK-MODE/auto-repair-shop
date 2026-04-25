package supply_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	infraPostgres "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	supplyRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/supply"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newSupply() *domain.Supply {
	return &domain.Supply{
		ID:            uuid.New().String(),
		Name:          "Óleo Motor 5W30",
		Description:   "Lubrificante sintético",
		UnitPrice:     decimal.NewFromFloat(49.90),
		StockQuantity: 100,
	}
}

func newListParams(page, pageSize int) *domain.ListSupplyParams {
	return &domain.ListSupplyParams{
		Page:     int64(page),
		PageSize: int64(pageSize),
	}
}

// setupMockDB injeta um *sql.DB mockado no singleton via ConnectWithDB.
// GetOneTimeTransaction chama Connect() que devolve esse mesmo *sql.DB.
// Para Save/Create usamos NewTransactionalUoW com esse mesmo db mockado.
func setupMockDB(t *testing.T) (sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	infraPostgres.ConnectWithDB(db)
	return mock, func() { db.Close() }
}

func supplyColumns() []string {
	return []string{"id", "name", "description", "unit_price", "stock_quantity", "version"}
}

// ---------------------------------------------------------------------------
// Repository()
// ---------------------------------------------------------------------------

func TestRepository_ReturnsNonNil(t *testing.T) {
	r := supplyRepo.Repository()
	assert.NotNil(t, r)
}

// ---------------------------------------------------------------------------
// Save
// ---------------------------------------------------------------------------

func TestSave_Success(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	supply := newSupply()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "supply"`).
		WithArgs(supply.ID, supply.Name, supply.Description, supply.UnitPrice, supply.StockQuantity, supply.Version).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	r := supplyRepo.Repository()
	uow := infraPostgres.NewTransactionalUoW(infraPostgres.Connect())

	err := uow.Execute(context.Background(), func(ctx context.Context) error {
		return r.Save(ctx, supply)
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSave_ExecError(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	supply := newSupply()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "supply"`).
		WithArgs(supply.ID, supply.Name, supply.Description, supply.UnitPrice, supply.StockQuantity, supply.Version).
		WillReturnError(errors.New("connection reset"))
	mock.ExpectRollback()

	r := supplyRepo.Repository()
	uow := infraPostgres.NewTransactionalUoW(infraPostgres.Connect())

	err := uow.Execute(context.Background(), func(ctx context.Context) error {
		return r.Save(ctx, supply)
	})

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSave_NoTransaction(t *testing.T) {
	r := supplyRepo.Repository()
	err := r.Save(context.Background(), newSupply())
	assert.ErrorIs(t, err, infraPostgres.ErrMissingPostgresTransaction)
}

func TestSave_ZeroValueSupply(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	supply := &domain.Supply{} // zero values

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "supply"`).
		WithArgs(supply.ID, supply.Name, supply.Description, supply.UnitPrice, supply.StockQuantity, supply.Version).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	r := supplyRepo.Repository()
	uow := infraPostgres.NewTransactionalUoW(infraPostgres.Connect())

	err := uow.Execute(context.Background(), func(ctx context.Context) error {
		return r.Save(ctx, supply)
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestCreate_Success(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	supply := newSupply()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "supply"`).
		WithArgs(supply.ID, supply.Name, supply.Description, supply.UnitPrice, supply.StockQuantity, supply.Version).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	r := supplyRepo.Repository()
	uow := infraPostgres.NewTransactionalUoW(infraPostgres.Connect())

	err := uow.Execute(context.Background(), func(ctx context.Context) error {
		return r.Save(ctx, supply)
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_ExecError(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	supply := newSupply()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "supply"`).
		WithArgs(supply.ID, supply.Name, supply.Description, supply.UnitPrice, supply.StockQuantity, supply.Version).
		WillReturnError(errors.New("unique violation"))
	mock.ExpectRollback()

	r := supplyRepo.Repository()
	uow := infraPostgres.NewTransactionalUoW(infraPostgres.Connect())

	err := uow.Execute(context.Background(), func(ctx context.Context) error {
		return r.Save(ctx, supply)
	})

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_NoTransaction(t *testing.T) {
	r := supplyRepo.Repository()
	err := r.Save(context.Background(), newSupply())
	assert.ErrorIs(t, err, infraPostgres.ErrMissingPostgresTransaction)
}

// ---------------------------------------------------------------------------
// DecrementStock
// ---------------------------------------------------------------------------

func TestDecrementStock_Success(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	supply := newSupply()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE supply`).
		WithArgs(supply.ID, 5).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	r := supplyRepo.Repository()
	uow := infraPostgres.NewTransactionalUoW(infraPostgres.Connect())

	err := uow.Execute(context.Background(), func(ctx context.Context) error {
		return r.DecrementStock(ctx, supply.ID, 5)
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDecrementStock_OutOfStock(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	supply := newSupply()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE supply`).
		WithArgs(supply.ID, 200).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows = condition not met
	mock.ExpectRollback()

	r := supplyRepo.Repository()
	uow := infraPostgres.NewTransactionalUoW(infraPostgres.Connect())

	err := uow.Execute(context.Background(), func(ctx context.Context) error {
		return r.DecrementStock(ctx, supply.ID, 200)
	})

	assert.ErrorIs(t, err, domain.ErrSupplyOutOfStock)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDecrementStock_ExecError(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	supply := newSupply()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE supply`).
		WithArgs(supply.ID, 1).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	r := supplyRepo.Repository()
	uow := infraPostgres.NewTransactionalUoW(infraPostgres.Connect())

	err := uow.Execute(context.Background(), func(ctx context.Context) error {
		return r.DecrementStock(ctx, supply.ID, 1)
	})

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDecrementStock_NoTransaction(t *testing.T) {
	r := supplyRepo.Repository()
	err := r.DecrementStock(context.Background(), "any-id", 1)
	assert.ErrorIs(t, err, infraPostgres.ErrMissingPostgresTransaction)
}

// ---------------------------------------------------------------------------
// RestoreStock
// ---------------------------------------------------------------------------

func TestRestoreStock_Success(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	supply := newSupply()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE supply`).
		WithArgs(supply.ID, 3).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	r := supplyRepo.Repository()
	uow := infraPostgres.NewTransactionalUoW(infraPostgres.Connect())

	err := uow.Execute(context.Background(), func(ctx context.Context) error {
		return r.RestoreStock(ctx, supply.ID, 3)
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRestoreStock_ExecError(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	supply := newSupply()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE supply`).
		WithArgs(supply.ID, 3).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	r := supplyRepo.Repository()
	uow := infraPostgres.NewTransactionalUoW(infraPostgres.Connect())

	err := uow.Execute(context.Background(), func(ctx context.Context) error {
		return r.RestoreStock(ctx, supply.ID, 3)
	})

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRestoreStock_NoTransaction(t *testing.T) {
	r := supplyRepo.Repository()
	err := r.RestoreStock(context.Background(), "any-id", 3)
	assert.ErrorIs(t, err, infraPostgres.ErrMissingPostgresTransaction)
}

// ---------------------------------------------------------------------------
// Count
// GetOneTimeTransaction retorna Connect() diretamente (não usa tx).
// Basta ter o mock injetado via ConnectWithDB.
// ---------------------------------------------------------------------------

func TestCount_Success(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(42))

	r := supplyRepo.Repository()
	total, err := r.Count(context.Background(), newListParams(1, 10))

	assert.NoError(t, err)
	assert.Equal(t, int64(42), total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCount_Zero(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	r := supplyRepo.Repository()
	total, err := r.Count(context.Background(), newListParams(1, 10))

	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCount_QueryError(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnError(errors.New("db error"))

	r := supplyRepo.Repository()
	total, err := r.Count(context.Background(), newListParams(1, 10))

	assert.Error(t, err)
	assert.Equal(t, int64(0), total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCount_NoRows(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnError(sql.ErrNoRows)

	r := supplyRepo.Repository()
	total, err := r.Count(context.Background(), newListParams(1, 10))

	assert.Error(t, err)
	assert.Equal(t, int64(0), total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

func TestSearch_Success(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	s := newSupply()

	mock.ExpectQuery(`SELECT s\.id`).
		WillReturnRows(
			sqlmock.NewRows(supplyColumns()).
				AddRow(s.ID, s.Name, s.Description, s.UnitPrice, s.StockQuantity, 1),
		)

	r := supplyRepo.Repository()
	results, err := r.Search(context.Background(), newListParams(1, 10))

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, s.ID, results[0].ID)
	assert.Equal(t, s.Name, results[0].Name)
	assert.Equal(t, s.Description, results[0].Description)
	assert.Equal(t, s.UnitPrice, results[0].UnitPrice)
	assert.Equal(t, s.StockQuantity, results[0].StockQuantity)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearch_MultipleRows(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	s1, s2 := newSupply(), newSupply()
	s2.Name = "Filtro de Ar"

	mock.ExpectQuery(`SELECT s\.id`).
		WillReturnRows(
			sqlmock.NewRows(supplyColumns()).
				AddRow(s1.ID, s1.Name, s1.Description, s1.UnitPrice, s1.StockQuantity, 1).
				AddRow(s2.ID, s2.Name, s2.Description, s2.UnitPrice, s2.StockQuantity, 2),
		)

	r := supplyRepo.Repository()
	results, err := r.Search(context.Background(), newListParams(1, 10))

	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearch_EmptyResult(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	mock.ExpectQuery(`SELECT s\.id`).
		WillReturnRows(sqlmock.NewRows(supplyColumns()))

	r := supplyRepo.Repository()
	results, err := r.Search(context.Background(), newListParams(1, 10))

	require.NoError(t, err)
	assert.Empty(t, results)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearch_QueryError(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	mock.ExpectQuery(`SELECT s\.id`).
		WillReturnError(errors.New("timeout"))

	r := supplyRepo.Repository()
	results, err := r.Search(context.Background(), newListParams(1, 10))

	assert.Error(t, err)
	assert.Nil(t, results)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearch_ScanError(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	// Uma coluna a menos força erro no Scan
	mock.ExpectQuery(`SELECT s\.id`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()),
		)

	r := supplyRepo.Repository()
	results, err := r.Search(context.Background(), newListParams(1, 10))

	assert.Error(t, err)
	assert.Nil(t, results)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearch_RowsError(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	s := newSupply()
	rowsErr := errors.New("row error during iteration")

	mock.ExpectQuery(`SELECT s\.id`).
		WillReturnRows(
			sqlmock.NewRows(supplyColumns()).
				AddRow(s.ID, s.Name, s.Description, s.UnitPrice, s.StockQuantity, 1).
				RowError(0, rowsErr),
		)

	r := supplyRepo.Repository()
	results, err := r.Search(context.Background(), newListParams(1, 10))

	assert.Error(t, err)
	assert.Nil(t, results)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearch_SecondPage(t *testing.T) {
	mock, close := setupMockDB(t)
	defer close()

	s := newSupply()

	mock.ExpectQuery(`SELECT s\.id`).
		WillReturnRows(
			sqlmock.NewRows(supplyColumns()).
				AddRow(s.ID, s.Name, s.Description, s.UnitPrice, s.StockQuantity, 1),
		)

	r := supplyRepo.Repository()
	// page=2, pageSize=5 → offset=5
	results, err := r.Search(context.Background(), newListParams(2, 5))

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_Supply_Count_NoTransaction(t *testing.T) {
	repo := supplyRepo.Repository()

	_, err := repo.Count(context.Background(), newListParams(1, 10))
	assert.Error(t, err)
}

func TestPostgresRepository_Delete_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := infraPostgres.NewTransactionalUoW(db)
	repo := supplyRepo.Repository()

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM \"supply\"").WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Delete(ctx, "id")
	})

	assert.Error(t, err)
}
func TestPostgresRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := infraPostgres.NewTransactionalUoW(db)
	repo := supplyRepo.Repository()

	id := "supply-id"

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM \"supply\"").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Delete(ctx, id)
	})

	assert.NoError(t, err)
}

func TestPostgresRepository_Delete_NoTransaction(t *testing.T) {
	repo := supplyRepo.Repository()

	err := repo.Delete(context.Background(), "some-id")
	assert.Error(t, err)
}
