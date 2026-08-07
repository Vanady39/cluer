package middlewares

import (
	"crypto/subtle"
	"net/http"
	"slices"
	"strings"

	"github.com/Vanady39/cluer/internal/domains"
	"github.com/gin-gonic/gin"
)

// Context keys for what the auth middlewares resolve.
const (
	ContextApp   = "app"
	ContextAppId = "app_id"
)

// AdminAuth guards every admin endpoint with a bearer token.
//
// Not wired into the router right now: authentication is deferred and the API
// runs as a single mocked user. Kept because it is the shape the real thing
// will take — one hardcoded administrator for the MVP — and because publishing,
// rollback and soft delete all live behind these routes, so they cannot stay
// open past local development.
func AdminAuth(token string) gin.HandlerFunc {
	expected := []byte(token)

	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		presented, ok := strings.CutPrefix(header, "Bearer ")
		// Constant-time compare: a byte-by-byte one leaks the token's prefix to
		// anyone willing to time the responses.
		if !ok || subtle.ConstantTimeCompare([]byte(presented), expected) != 1 {
			c.Error(&AuthHeaderError{
				Code: http.StatusUnauthorized,
				msg:  "Valid admin bearer token required",
				Err:  errUnauthorized,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// AppKeyAuth resolves the consumer application from its public key and applies
// that application's CORS policy.
//
// This is not user authentication and is unaffected by deferring it: the key
// identifies which application is calling, which /resolve needs in order to know
// whose tours to look for at all.
//
// The key is not a secret — it sits in the page source — so it grants exactly
// two things: read a published tour and post events. The origin allowlist is
// what stops someone else's site from filling your analytics with noise using a
// key they read off your page.
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

// RuntimeCORS answers the preflight before the key is checked, because a
// browser sends OPTIONS without custom headers and would otherwise never get to
// send the real request.
func RuntimeCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodOptions {
			c.Next()
			return
		}

		if origin := c.GetHeader("Origin"); origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, X-App-Key")
		c.Header("Access-Control-Max-Age", "600")
		c.AbortWithStatus(http.StatusNoContent)
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
