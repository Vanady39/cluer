package controllers_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Vanady39/cluer/internal/controllers"
	"github.com/Vanady39/cluer/internal/domains"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubUserDomain struct {
	user domains.User
	err  error
}

func (s *stubUserDomain) GetCurrentUser(_ context.Context) (domains.User, error) {
	if s.err != nil {
		return domains.User{}, s.err
	}
	return s.user, nil
}

func userRouter(domain domains.UserDomainInterface) *gin.Engine {
	controller := controllers.NewUserController(domain)
	return newTestRouter(func(r *gin.Engine) {
		r.GET("/users/me", controller.GetCurrentUser)
	})
}

func TestUserController_GetCurrentUser(t *testing.T) {
	t.Run("returns the user wrapped in a data envelope", func(t *testing.T) {
		domain := &stubUserDomain{user: domains.User{
			ID: 1, Name: "Иванов Иван", AvatarURL: "https://x/a.png",
		}}
		rec := doJSON(t, userRouter(domain), http.MethodGet, "/users/me", nil)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t,
			`{"data":{"id":1,"name":"Иванов Иван","avatarUrl":"https://x/a.png"}}`,
			rec.Body.String(),
		)
	})

	t.Run("domain failure returns 500 in the standard envelope", func(t *testing.T) {
		rec := doJSON(t, userRouter(&stubUserDomain{err: assert.AnError}), http.MethodGet, "/users/me", nil)

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		decodeHTTPError(t, rec)
	})
}
