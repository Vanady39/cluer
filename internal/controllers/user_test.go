package controllers_test

import (
	"net/http"
	"testing"

	"github.com/Vanady39/cluer/internal/auth"
	"github.com/Vanady39/cluer/internal/controllers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// userRouter mounts the handler behind a stand-in for OIDCAuth: the real
// middleware needs a verifier and a signed token, while all this controller
// sees of it is the encoded claims it leaves in the context. Passing nil claims
// mounts the route bare, which is how the "no claims" case is reached.
func userRouter(t *testing.T, claims *auth.Claims) *gin.Engine {
	t.Helper()

	controller := controllers.NewUserController()
	return newTestRouter(func(r *gin.Engine) {
		handlers := []gin.HandlerFunc{}
		if claims != nil {
			encoded, err := auth.EncodeGob(*claims)
			require.NoError(t, err)
			handlers = append(handlers, func(c *gin.Context) {
				c.Set(auth.ContextClaims, encoded)
			})
		}
		handlers = append(handlers, controller.GetCurrentUser)

		r.GET("/users/me", handlers...)
	})
}

func TestUserController_GetCurrentUser(t *testing.T) {
	t.Run("returns the identity from the token in a data envelope", func(t *testing.T) {
		claims := auth.Claims{
			Subject:           "ak-visitor-1",
			Email:             "visitor@example.com",
			Name:              "Иванов Иван",
			PreferredUsername: "ivanov",
			Picture:           "https://x/a.png",
			// Groups are carried by the token but say nothing on the
			// storefront, so they must not leak into its response.
			Groups: []string{"cluer-admins"},
		}
		rec := doJSON(t, userRouter(t, &claims), http.MethodGet, "/users/me", nil)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t,
			`{"data":{"subject":"ak-visitor-1","email":"visitor@example.com",`+
				`"name":"Иванов Иван","username":"ivanov","avatarUrl":"https://x/a.png"}}`,
			rec.Body.String(),
		)
	})

	// An IdP that releases nothing but the subject still produces a valid
	// answer: the profile fields come back empty rather than missing, so the
	// client always parses the same shape.
	t.Run("profile claims the IdP withheld come back empty", func(t *testing.T) {
		rec := doJSON(t, userRouter(t, &auth.Claims{Subject: "ak-visitor-2"}), http.MethodGet, "/users/me", nil)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t,
			`{"data":{"subject":"ak-visitor-2","email":"","name":"","username":"","avatarUrl":""}}`,
			rec.Body.String(),
		)
	})

	// Reachable only by wiring the route outside the authenticated chain. It
	// answers 401 rather than an anonymous user, which is the whole point of
	// removing the mock.
	t.Run("without claims in the context returns 401", func(t *testing.T) {
		rec := doJSON(t, userRouter(t, nil), http.MethodGet, "/users/me", nil)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
		decodeHTTPError(t, rec)
	})
}
