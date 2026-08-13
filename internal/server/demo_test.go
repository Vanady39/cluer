package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Vanady39/cluer/internal/config"
	"github.com/Vanady39/cluer/internal/controllers"
	"github.com/Vanady39/cluer/internal/domains"
	"github.com/Vanady39/cluer/internal/models/response"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	demoIssuer   = "https://auth.example.com/application/o/cluer/"
	demoClientID = "cluer-admin"
	demoSubject  = "ak-visitor-1"
)

type demoListingControllerStub struct{}

func (demoListingControllerStub) GetListings(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, response.NewGetListingsResponse(domains.ListingPage{}))
}

type echoKeySet struct{}

func (echoKeySet) VerifySignature(_ context.Context, jwt string) ([]byte, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return nil, errors.New("malformed jwt")
	}
	return base64.RawURLEncoding.DecodeString(parts[1])
}

func demoToken(t *testing.T, claims map[string]any) string {
	t.Helper()

	payload, err := json.Marshal(claims)
	require.NoError(t, err)

	return strings.Join([]string{
		base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256"}`)),
		base64.RawURLEncoding.EncodeToString(payload),
		base64.RawURLEncoding.EncodeToString([]byte("signature")),
	}, ".")
}

func newDemoTestServer(t *testing.T) *Server {
	t.Helper()

	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard

	logger := zerolog.Nop()
	srv := NewDemoServer(&config.ServerConfig{Address: "127.0.0.1:0"}, &DemoCreateStruct{
		Logger:            &logger,
		OIDCVerifier:      oidc.NewVerifier(demoIssuer, echoKeySet{}, &oidc.Config{ClientID: demoClientID}),
		ListingController: demoListingControllerStub{},
		UserController:    controllers.NewUserController(),
	})
	require.NotNil(t, srv.httpServer)

	return srv
}

func demoCall(srv *Server, method, target, authorization string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}

	rec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(rec, req)

	return rec
}

// TestNewDemoServerRoutes фиксирует то, что описано в listings-api.md:
// у demo-сервиса ровно два роута под /v1.
func TestNewDemoServerRoutes(t *testing.T) {
	srv := newDemoTestServer(t)

	call := func(method, target string) int {
		return demoCall(srv, method, target, "").Code
	}

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

// Лента — витрина: её открывают посетители без токена, и аутентификация,
// навешанная на группу вместо конкретного роута, сломала бы именно это.
func TestDemoListingsStayPublic(t *testing.T) {
	rec := demoCall(newDemoTestServer(t), http.MethodGet, "/v1/listings", "")

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestDemoCurrentUserRequiresToken(t *testing.T) {
	srv := newDemoTestServer(t)

	tests := []struct {
		name          string
		authorization func(t *testing.T) string
	}{
		{name: "без заголовка", authorization: func(*testing.T) string { return "" }},
		{name: "не токен", authorization: func(*testing.T) string { return "Bearer not-a-token" }},
		{
			name: "просроченный",
			authorization: func(t *testing.T) string {
				claims := demoClaims()
				claims["exp"] = time.Now().Add(-time.Minute).Unix()
				return "Bearer " + demoToken(t, claims)
			},
		},
		{
			name: "выдан другим провайдером",
			authorization: func(t *testing.T) string {
				claims := demoClaims()
				claims["iss"] = "https://auth.attacker.example/application/o/cluer/"
				return "Bearer " + demoToken(t, claims)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := demoCall(srv, http.MethodGet, "/v1/users/me", tt.authorization(t))

			require.Equal(t, http.StatusUnauthorized, rec.Code)
		})
	}
}

// Главное отличие витрины от админки: групп здесь не спрашивают, и токен без
// единой группы обязан получить 200, а не 403.
func TestDemoCurrentUserAcceptsAnyAuthenticatedVisitor(t *testing.T) {
	rec := demoCall(newDemoTestServer(t), http.MethodGet, "/v1/users/me", "Bearer "+demoToken(t, demoClaims()))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t,
		`{"data":{"subject":"`+demoSubject+`","email":"visitor@example.com",`+
			`"name":"Иванов Иван","username":"ivanov","avatarUrl":"https://example.com/avatar.png"}}`,
		rec.Body.String(),
	)
}

func demoClaims() map[string]any {
	return map[string]any{
		"iss":                demoIssuer,
		"aud":                demoClientID,
		"sub":                demoSubject,
		"exp":                time.Now().Add(time.Hour).Unix(),
		"iat":                time.Now().Unix(),
		"email":              "visitor@example.com",
		"email_verified":     true,
		"name":               "Иванов Иван",
		"preferred_username": "ivanov",
		"picture":            "https://example.com/avatar.png",
	}
}
