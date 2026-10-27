package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"news-portal/apps/api/internal/pkg/response"
	"news-portal/apps/api/internal/transport/http/handler"
)

func TestHealthHandler_Healthz(t *testing.T) {
	h := handler.NewHealthHandler(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	h.Healthz(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

	var env response.ResponseEnvelope
	err := json.Unmarshal(rec.Body.Bytes(), &env)
	require.NoError(t, err)

	assert.True(t, env.Success)
	dataMap, ok := env.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "ok", dataMap["status"])
}

func TestHealthHandler_Readyz_WithoutServices(t *testing.T) {
	h := handler.NewHealthHandler(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	h.Readyz(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var env response.ResponseEnvelope
	err := json.Unmarshal(rec.Body.Bytes(), &env)
	require.NoError(t, err)

	assert.True(t, env.Success)
}
