// Package controllers_test is an external test package: the tests install the real
// ErrorHandler middleware, and middlewares imports controllers, so testing from
// inside package controllers would be an import cycle.
package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/Vanady39/cluer/internal/models"
	"github.com/Vanady39/cluer/internal/server/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestRouter builds a router carrying the same ErrorHandler the real server
// installs, so tests exercise the actual error-rendering path rather than
// assuming what a controller would have written.
func newTestRouter(register func(r *gin.Engine)) *gin.Engine {
	logger := zerolog.Nop()
	router := gin.New()
	router.Use(middlewares.ErrorHandler(&logger))
	register(router)
	return router
}

func doJSON(t *testing.T, router *gin.Engine, method, target string, body any) *httptest.ResponseRecorder {
	t.Helper()

	raw := []byte(nil)
	if body != nil {
		encoded, err := json.Marshal(body)
		require.NoError(t, err)
		raw = encoded
	}

	req := httptest.NewRequest(method, target, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// doRaw sends a body that is not necessarily valid JSON.
func doRaw(router *gin.Engine, method, target, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// decodeHTTPError asserts the body is the standard error envelope and returns it.
func decodeHTTPError(t *testing.T, rec *httptest.ResponseRecorder) models.HTTPError {
	t.Helper()

	var payload models.HTTPError
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.NotEmpty(t, payload.Error, "error field must be populated")
	require.NotEmpty(t, payload.Message, "message field must be populated")
	return payload
}
