package middlewares

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/Vanady39/cluer/platform/oidcauth"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
)

const authorizationHeader = "Authorization"

func OIDCAuth(verifier *oidc.IDTokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader(authorizationHeader)
		if header == "" {
			c.Error(&AuthHeaderError{
				Code: http.StatusUnauthorized,
				msg:  "Authorization header is required",
				Err:  errMissingAuthHeader,
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.Error(&AuthHeaderError{
				Code: http.StatusUnauthorized,
				msg:  "Authorization header must be 'Bearer <token>'",
				Err:  errMalformedAuthHeader,
			})
			c.Abort()
			return
		}

		rawIDToken := strings.TrimSpace(parts[1])
		if rawIDToken == "" {
			c.Error(&AuthHeaderError{
				Code: http.StatusUnauthorized,
				msg:  "Authorization header must be 'Bearer <token>'",
				Err:  errMalformedAuthHeader,
			})
			c.Abort()
			return
		}

		idToken, err := verifier.Verify(c.Request.Context(), rawIDToken)
		if err != nil {
			c.Error(&AuthHeaderError{
				Code: http.StatusUnauthorized,
				msg:  "Token is invalid or expired",
				Err:  fmt.Errorf("%w: %v", errInvalidToken, err),
			})
			c.Abort()
			return
		}

		var claims oidcauth.Claims
		if err := idToken.Claims(&claims); err != nil {
			c.Error(&AuthHeaderError{
				Code: http.StatusUnauthorized,
				msg:  "Token is invalid or expired",
				Err:  fmt.Errorf("%w: %v", errInvalidToken, err),
			})
			c.Abort()
			return
		}

		encoded, err := oidcauth.EncodeGob(claims)
		if err != nil {
			c.Error(&AuthHeaderError{
				Code: http.StatusInternalServerError,
				msg:  "Authentication context is broken",
				Err:  fmt.Errorf("%w: %v", errBrokenAuthContext, err),
			})
			c.Abort()
			return
		}

		c.Set(oidcauth.ContextClaims, encoded)
		c.Next()
	}
}

func RequireGroups(groups []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := oidcauth.FromGin(c)
		if err != nil {
			c.Error(&AuthHeaderError{
				Code: http.StatusInternalServerError,
				msg:  "Authentication context is broken",
				Err:  fmt.Errorf("%w: %v", errBrokenAuthContext, err),
			})
			c.Abort()
			return
		}

		for _, group := range claims.Groups {
			if slices.Contains(groups, group) {
				c.Next()
				return
			}
		}

		c.Error(&AuthHeaderError{
			Code: http.StatusForbidden,
			msg:  "Access is limited to administrators",
			Err:  errInsufficientGroups,
		})
		c.Abort()
	}
}
