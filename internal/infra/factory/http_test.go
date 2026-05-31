package factory

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	postgresdb "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	"github.com/stretchr/testify/assert"
)

func TestHttpContainer(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	container := httpContainer(db)

	assert.NotNil(t, container)
	assert.NotNil(t, container.PingHandler)
	assert.NotNil(t, container.UserHandler)
	assert.NotNil(t, container.CustomerHandler)
	assert.NotNil(t, container.WorkHandler)
	assert.NotNil(t, container.SupplyHandler)

}

func TestMiddlewaresContainer(t *testing.T) {
	m := middlewaresContainer()
	assert.NotNil(t, m)
	assert.Contains(t, *m, "Logger")
	assert.Contains(t, *m, "Recovery")
	assert.Contains(t, *m, "ErrorHandler")
}

func TestHTTPServer(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)

	s := HTTPServer()
	assert.NotNil(t, s)
}
