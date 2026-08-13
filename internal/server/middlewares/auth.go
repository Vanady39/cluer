package middlewares

import (
	"net/http"
	"slices"
	"strings"

	"github.com/Vanady39/cluer/internal/domains"
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
			c.Error(&AuthHeaderError{
				Code: http.StatusUnauthorized,
				msg:  "X-App-Key header is required",
				Err:  domains.ErrAppKeyRequired,
			})
			c.Abort()
			return
		}

		app, err := runtime.AppByKey(c.Request.Context(), key)
		if err != nil {
			c.Error(&AuthHeaderError{
				Code: http.StatusUnauthorized,
				msg:  "Unknown application key",
				Err:  domains.ErrAppNotFound,
			})
			c.Abort()
			return
		}

		origin := c.GetHeader("Origin")
		if origin != "" {
			if !originAllowed(app.AllowedOrigins, origin) {
				c.Error(&AuthHeaderError{
					Code: http.StatusForbidden,
					msg:  "Origin is not allowed for this application",
					Err:  domains.ErrOriginNotAllowed,
				})
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

func RuntimeCORS() gin.HandlerFunc {
	return func(c *gin.Context) {

		if origin := c.GetHeader("Origin"); origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}

		c.Header(
			"Access-Control-Allow-Methods",
			"GET, POST, PATCH, PUT, DELETE, OPTIONS",
		)

		c.Header(
			"Access-Control-Allow-Headers",
			"Content-Type, X-App-Key, Authorization",
		)

		c.Header(
			"Access-Control-Expose-Headers",
			"Location",
		)

		c.Header(
			"Access-Control-Max-Age",
			"600",
		)

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

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
