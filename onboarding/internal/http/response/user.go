package response

import (
	"github.com/Vanady39/cluer/platform/oidcauth"
)

type CurrentAdmin struct {
	Subject  string   `json:"subject"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	Username string   `json:"username"`
	Groups   []string `json:"groups"`
}

type GetCurrentAdminResponse struct {
	Data CurrentAdmin `json:"data"`
}

func NewGetCurrentAdminResponse(claims oidcauth.Claims) GetCurrentAdminResponse {
	groups := claims.Groups
	if groups == nil {
		groups = []string{}
	}

	return GetCurrentAdminResponse{
		Data: CurrentAdmin{
			Subject:  claims.Subject,
			Email:    claims.Email,
			Name:     claims.Name,
			Username: claims.PreferredUsername,
			Groups:   groups,
		},
	}
}
