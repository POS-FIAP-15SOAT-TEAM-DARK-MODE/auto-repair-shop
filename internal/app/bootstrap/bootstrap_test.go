package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPServer_BuildsGatewayFromEnvironment(t *testing.T) {
	t.Setenv("PORT", "18080")
	t.Setenv("UPSTREAM_TIMEOUT", "10s")
	t.Setenv("LOG_LEVEL", "PANIC")

	srv, err := HTTPServer()

	require.NoError(t, err)
	assert.Equal(t, "gateway", srv.Name())
}

func TestHTTPServer_FailsOnInvalidConfiguration(t *testing.T) {
	t.Setenv("UPSTREAM_TIMEOUT", "0s")

	srv, err := HTTPServer()

	assert.Error(t, err)
	assert.Nil(t, srv)
}
