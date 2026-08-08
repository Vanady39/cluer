package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAdminAuthRejectsUnusableToken(t *testing.T) {
	for _, token := range []string{"", " ", "\t\n"} {
		handler, err := AdminAuth(token)
		assert.ErrorIs(t, err, ErrEmptyAdminToken, "token %q", token)
		assert.Nil(t, handler, "token %q", token)
	}
}

func TestAdminAuth(t *testing.T) {
	const token = "a-token-long-enough-to-be-plausible"

	handler, err := AdminAuth(token)
	require.NoError(t, err)

	logger := zerolog.Nop()
	router := gin.New()
	router.Use(ErrorHandler(&logger), handler)
	router.GET("/guarded", func(c *gin.Context) { c.Status(http.StatusOK) })

	tests := []struct {
		name   string
		header string
		want   int
	}{
		{"верный токен", "Bearer " + token, http.StatusOK},
		{"нет заголовка", "", http.StatusUnauthorized},
		{"чужой токен", "Bearer nope", http.StatusUnauthorized},
		{"пустой после префикса", "Bearer ", http.StatusUnauthorized},
		{"только пробелы после префикса", "Bearer    ", http.StatusUnauthorized},
		{"префикс без пробела", "Bearer", http.StatusUnauthorized},
		{"без схемы", token, http.StatusUnauthorized},
		{"другая схема", "Basic " + token, http.StatusUnauthorized},
		{"верный токен с лишними пробелами", "Bearer   " + token + "  ", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/guarded", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.want, rec.Code)
		})
	}
}
