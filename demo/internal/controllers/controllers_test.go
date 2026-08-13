package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/Vanady39/cluer/platform/errs"
	"github.com/Vanady39/cluer/platform/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

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

func decodeHTTPError(t *testing.T, rec *httptest.ResponseRecorder) errs.HTTPError {
	t.Helper()

	var payload errs.HTTPError
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.NotEmpty(t, payload.Error, "error field must be populated")
	require.NotEmpty(t, payload.Message, "message field must be populated")
	return payload
}
