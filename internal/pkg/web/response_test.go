package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteError_UsesEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteError(rec, http.StatusBadGateway, "upstream unavailable")

	assert.Equal(t, http.StatusBadGateway, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"code":502,"errors":["upstream unavailable"]}`, rec.Body.String())
}

func TestWriteJSON_WritesStatusAndBody(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteJSON(rec, http.StatusAccepted, map[string]string{"message": "ok"})

	assert.Equal(t, http.StatusAccepted, rec.Code)
	assert.JSONEq(t, `{"message":"ok"}`, rec.Body.String())
}
