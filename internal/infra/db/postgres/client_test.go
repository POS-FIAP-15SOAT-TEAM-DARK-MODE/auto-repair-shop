package postgres

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestConnect_ReturnsExistingInstance(t *testing.T) {
	db, _, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	instance = nil
	once = newOnce()

	ConnectWithDB(db)

	got := Connect()
	assert.NotNil(t, got)
	assert.Equal(t, db, got)
}

func TestConnect_UsesOpenFnAndPings(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectPing()

	instance = nil
	once = newOnce()
	restore := setSQLOpenFn(func(driverName, dataSourceName string) (*sql.DB, error) {
		return db, nil
	})
	defer restore()

	got := Connect()
	assert.NotNil(t, got)
	assert.Equal(t, db, got)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestConnect_OpenFailure(t *testing.T) {
	if os.Getenv("TEST_CONNECT_OPEN_FAILURE") == "1" {
		instance = nil
		once = newOnce()
		setSQLOpenFn(func(driverName, dataSourceName string) (*sql.DB, error) {
			return nil, fmt.Errorf("open failed")
		})
		Connect()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestConnect_OpenFailure")
	cmd.Env = append(os.Environ(), "TEST_CONNECT_OPEN_FAILURE=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatal("expected process to exit with failure")
}

func TestConnect_PingFailure(t *testing.T) {
	if os.Getenv("TEST_CONNECT_PING_FAILURE") == "1" {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		if err != nil {
			os.Exit(1)
		}
		defer func() { _ = db.Close() }()
		mock.ExpectPing().WillReturnError(fmt.Errorf("ping failed"))
		instance = nil
		once = newOnce()
		setSQLOpenFn(func(driverName, dataSourceName string) (*sql.DB, error) {
			return db, nil
		})
		Connect()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestConnect_PingFailure")
	cmd.Env = append(os.Environ(), "TEST_CONNECT_PING_FAILURE=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatal("expected process to exit with failure")
}

func TestConnectWithDB_SetsInstance(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	instance = nil
	once = newOnce()

	ConnectWithDB(db)

	assert.Equal(t, db, instance)
}

func TestConnectWithDB_OnlyFirstCallWins(t *testing.T) {
	db1, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db1.Close() }()

	db2, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db2.Close() }()

	instance = nil
	once = newOnce()

	ConnectWithDB(db1)
	ConnectWithDB(db2)

	assert.Equal(t, db1, instance)
}

func newOnce() sync.Once {
	return sync.Once{}
}

func setSQLOpenFn(fn func(driverName, dataSourceName string) (*sql.DB, error)) func() {
	orig := sqlOpenFn
	sqlOpenFn = fn
	return func() { sqlOpenFn = orig }
}
