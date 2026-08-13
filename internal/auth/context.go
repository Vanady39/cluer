package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
)

const ContextClaims = "OIDCClaims"

var ErrNoClaims = errors.New("oidc claims are missing from the context")

func FromGin(c *gin.Context) (Claims, error) {
	raw := c.GetString(ContextClaims)
	if raw == "" {
		return Claims{}, ErrNoClaims
	}

	var claims Claims
	if err := DecodeGob(raw, &claims); err != nil {
		return Claims{}, err
	}

	return claims, nil
}
