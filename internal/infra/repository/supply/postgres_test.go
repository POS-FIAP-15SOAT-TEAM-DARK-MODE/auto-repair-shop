package supply_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	postgresdb "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	supplyrepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/supply"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// fullSupplyRepo expõe todos os métodos do repositório concreto,
// incluindo os que não fazem parte da interface domain.SupplyRepository.
type fullSupplyRepo interface {
	domain.SupplyRepository
	Create(context.Context, *domain.Supply) error
	List(context.Context) ([]*domain.Supply, error)
	Update(context.Context, *domain.SupplyUpdate) error
}

func newRepo() fullSupplyRepo {
	return supplyrepo.Repository().(fullSupplyRepo)
}

func newTestSupply() *domain.Supply {
	return domain.NewSupply("Oil Filter", "A good oil filter description", decimal.NewFromFloat(25.50), 10, 0)
}

// ---------------------------------------------------------------------------
// Save
// ---------------------------------------------------------------------------

func TestPostgresRepository_Supply_Save(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := newRepo()

	s := newTestSupply()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "supply"`).
		WithArgs(s.ID, s.Name, s.Description, s.UnitPrice, s.StockQuantity).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Save(ctx, s)
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_Supply_Save_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := newRepo()

	s := newTestSupply()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "supply"`).WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Save(ctx, s)
	})

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_Supply_Save_NoTransaction(t *testing.T) {
	repo := newRepo()
	s := newTestSupply()

	err := repo.Save(context.Background(), s)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestPostgresRepository_Supply_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := newRepo()

	s := newTestSupply()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "supply"`).
		WithArgs(s.ID, s.Name, s.Description, s.UnitPrice, s.StockQuantity, s.Version).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Create(ctx, s)
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_Supply_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := newRepo()

	s := newTestSupply()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "supply"`).WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Create(ctx, s)
	})

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_Supply_Create_NoTransaction(t *testing.T) {
	repo := newRepo()
	s := newTestSupply()

	err := repo.Create(context.Background(), s)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestPostgresRepository_Supply_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := newRepo()

	mock.ExpectBegin()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "unit_price", "stock_quantity", "version"}).
		AddRow("id1", "Supply 1", "Description one", decimal.NewFromFloat(10.00), 5, 0)
	mock.ExpectQuery(`SELECT`).WillReturnRows(rows)
	mock.ExpectCommit()

	var results []*domain.Supply
	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		var err error
		results, err = repo.List(ctx)
		return err
	})

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "id1", results[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_Supply_List_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := newRepo()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT`).WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		_, err := repo.List(ctx)
		return err
	})

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_Supply_List_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := newRepo()

	mock.ExpectBegin()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "unit_price", "stock_quantity", "version"}).
		AddRow("id1", "Supply 1", "Description one", "not-a-decimal", "not-an-int", "not-an-int")
	mock.ExpectQuery(`SELECT`).WillReturnRows(rows)
	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		_, err := repo.List(ctx)
		return err
	})

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_Supply_List_NoTransaction(t *testing.T) {
	repo := newRepo()

	_, err := repo.List(context.Background())
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestPostgresRepository_Supply_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := newRepo()

	name := "Updated Name"
	update := domain.UpdateSupply("supply-id", &name, nil, nil, nil)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "supply"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Update(ctx, update)
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_Supply_Update_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := newRepo()

	name := "Updated Name"
	update := domain.UpdateSupply("nonexistent-id", &name, nil, nil, nil)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "supply"`).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Update(ctx, update)
	})

	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_Supply_Update_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := newRepo()

	name := "Updated Name"
	update := domain.UpdateSupply("supply-id", &name, nil, nil, nil)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "supply"`).WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Update(ctx, update)
	})

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_Supply_Update_NoTransaction(t *testing.T) {
	repo := newRepo()
	name := "Updated Name"
	update := domain.UpdateSupply("supply-id", &name, nil, nil, nil)

	err := repo.Update(context.Background(), update)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

func TestPostgresRepository_Supply_Search(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := newRepo()

	params := &domain.SearchSupplyParams{Limit: 10, Offset: 0}

	mock.ExpectBegin()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "unit_price"}).
		AddRow("id1", "Supply 1", "Description one", decimal.NewFromFloat(10.00))
	mock.ExpectQuery(`SELECT s.id, s.name, s.description, s.unit_price`).
		WillReturnRows(rows)
	mock.ExpectCommit()

	var results []domain.Supply
	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		var err error
		results, err = repo.Search(ctx, params)
		return err
	})

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "id1", results[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_Supply_Search_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := newRepo()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT s.id, s.name, s.description, s.unit_price`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		_, err := repo.Search(ctx, &domain.SearchSupplyParams{Limit: 10})
		return err
	})

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_Supply_Search_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := newRepo()

	mock.ExpectBegin()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "unit_price"}).
		AddRow("id1", "Supply 1", "Description one", "not-a-decimal") // scan error
	mock.ExpectQuery(`SELECT s.id, s.name, s.description, s.unit_price`).
		WillReturnRows(rows)
	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		_, err := repo.Search(ctx, &domain.SearchSupplyParams{Limit: 10})
		return err
	})

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_Supply_Search_NoTransaction(t *testing.T) {
	repo := newRepo()

	_, err := repo.Search(context.Background(), &domain.SearchSupplyParams{Limit: 10})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Count
// ---------------------------------------------------------------------------

func TestPostgresRepository_Supply_Count(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := newRepo()

	params := &domain.SearchSupplyParams{}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT COUNT\(s.id\) FROM "supply"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))
	mock.ExpectCommit()

	var total int64
	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		var err error
		total, err = repo.Count(ctx, params)
		return err
	})

	assert.NoError(t, err)
	assert.Equal(t, int64(7), total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_Supply_Count_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := newRepo()

	params := &domain.SearchSupplyParams{}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT COUNT\(s.id\) FROM "supply"`).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	var total int64
	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		var err error
		total, err = repo.Count(ctx, params)
		return err
	})

	assert.Error(t, err)
	assert.Equal(t, int64(0), total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_Supply_Count_WithStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := newRepo()

	params := &domain.SearchSupplyParams{Status: "ACTIVE"}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT COUNT\(s.id\) FROM "supply"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectCommit()

	var total int64
	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		var err error
		total, err = repo.Count(ctx, params)
		return err
	})

	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_Supply_Count_NoTransaction(t *testing.T) {
	repo := newRepo()

	_, err := repo.Count(context.Background(), &domain.SearchSupplyParams{})
	assert.Error(t, err)
}
