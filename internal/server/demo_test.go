package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vanady39/cluer/internal/config"
	"github.com/Vanady39/cluer/internal/domains"
	"github.com/Vanady39/cluer/internal/models/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type demoListingControllerStub struct{}

func (demoListingControllerStub) GetListings(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, response.NewGetListingsResponse(domains.ListingPage{}))
}

type demoUserControllerStub struct{}

func (demoUserControllerStub) GetCurrentUser(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, response.NewGetCurrentUserResponse(domains.User{}))
}

// TestNewDemoServerRoutes фиксирует то, что описано в listings-api.md:
// у demo-сервиса ровно два роута под /v1.
func TestNewDemoServerRoutes(t *testing.T) {
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard

	logger := zerolog.Nop()
	srv := NewDemoServer(&config.ServerConfig{Address: "127.0.0.1:0"}, &DemoCreateStruct{
		Logger:            &logger,
		ListingController: demoListingControllerStub{},
		UserController:    demoUserControllerStub{},
	})
	require.NotNil(t, srv.httpServer)
	handler := srv.httpServer.Handler

	call := func(method, target string) int {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
		return rec.Code
	}

	t.Run("оба роута контракта доступны", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, call(http.MethodGet, "/v1/listings"))
		assert.Equal(t, http.StatusOK, call(http.MethodGet, "/v1/users/me"))
	})

	t.Run("других роутов под /v1 нет", func(t *testing.T) {
		for _, target := range []string{
			"/v1/health",
			"/v1/listings/1",
			"/v1/users",
			"/v1/tours",
			"/v1",
		} {
			assert.Equal(t, http.StatusNotFound, call(http.MethodGet, target), "неожиданный роут %s", target)
		}
	})

	t.Run("listings отвечает только на GET", func(t *testing.T) {
		for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
			assert.NotEqual(t, http.StatusOK, call(method, "/v1/listings"), "метод %s не описан в контракте", method)
		}
	})

	t.Run("shutdown не возвращает ошибку", func(t *testing.T) {
		require.NoError(t, srv.Shutdown(context.Background()))
	})
}
