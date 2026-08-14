// Package middlewares holds the HTTP middleware that only the onboarding
// service can have: AppKeyAuth resolves an application from its public key,
// which needs the runtime domain. That dependency is exactly why it cannot sit
// in platform/middlewares alongside the OIDC and CORS handlers.
package middlewares

import (
	"net/http"
	"slices"
	"strings"

	"github.com/Vanady39/cluer/onboarding/internal/domains"
	platform "github.com/Vanady39/cluer/platform/middlewares"
	"github.com/gin-gonic/gin"
)

const (
	ContextApp   = "app"
	ContextAppId = "app_id"
)

func AppKeyAuth(runtime domains.RuntimeDomainInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-App-Key")
		if key == "" {
			c.Error(platform.NewAuthHeaderError(
				http.StatusUnauthorized,
				"X-App-Key header is required",
				domains.ErrAppKeyRequired,
			))
			c.Abort()
			return
		}

		app, err := runtime.AppByKey(c.Request.Context(), key)
		if err != nil {
			c.Error(platform.NewAuthHeaderError(
				http.StatusUnauthorized,
				"Unknown application key",
				domains.ErrAppNotFound,
			))
			c.Abort()
			return
		}

		origin := c.GetHeader("Origin")
		if origin != "" {
			if !domains.OriginAllowed(app.AllowedOrigins, origin) {
				c.Error(platform.NewAuthHeaderError(
					http.StatusForbidden,
					"Origin is not allowed for this application",
					domains.ErrOriginNotAllowed,
				))
				c.Abort()
				return
			}
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}

		c.Set(ContextApp, app)
		c.Set(ContextAppId, app.Id)
		c.Next()
	}
}

func originAllowed(allowed []string, origin string) bool {
	if len(allowed) == 0 {
		return false
	}
	if slices.Contains(allowed, "*") {
		return true
	}
	return slices.Contains(allowed, strings.TrimSuffix(origin, "/"))
}
