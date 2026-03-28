package web_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/json"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
	"github.com/stretchr/testify/assert"
)

type joinedWithNil struct{}

func (joinedWithNil) Error() string   { return "joined with nil" }
func (joinedWithNil) Unwrap() []error { return []error{nil, errors.New("real error")} }

func TestError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{"BadRequest_UserName", domain.ErrEmptyUserName, http.StatusBadRequest},
		{"BadRequest_Joined", errors.Join(errors.New("err1"), errors.New("err2")), http.StatusBadRequest},
		{"InternalServer_JSONUnsupported", json.ErrJSONUnsupportedValue, http.StatusInternalServerError},
		{"Conflict", domain.ErrDataConflict, http.StatusConflict},
		{"UnprocessableEntity", domain.ErrWorkPriceLessThenOrEqualZero, http.StatusUnprocessableEntity},
		{"Default_InternalServer", errors.New("random error"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp := web.Error(tt.err)
			assert.Equal(t, tt.expectedStatus, status)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestIsJoined(t *testing.T) {
	err := errors.Join(errors.New("a"), errors.New("b"))
	assert.True(t, web.IsJoined(err))

	err = errors.New("single")
	assert.False(t, web.IsJoined(err))
}

func TestBuildErrorMessages(t *testing.T) {
	status, resp := web.Error(domain.ErrEmptyUserName)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, resp.Errors, domain.ErrEmptyUserName.Error())

	status, resp = web.Error(errors.Join(domain.ErrEmptyUserName, domain.ErrEmptyUserEmail))
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Len(t, resp.Errors, 2)

	status, resp = web.Error(domain.ErrDataConflict)
	assert.Equal(t, http.StatusConflict, status)
	assert.Equal(t, domain.ErrDataConflict.Error(), resp.Error)

	status, resp = web.Error(errors.New("fatal"))
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.Equal(t, "fatal", resp.Error)

	status, resp = web.Error(joinedWithNil{})
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, resp.Errors, "real error")
}
