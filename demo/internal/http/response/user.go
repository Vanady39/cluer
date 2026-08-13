package response

import (
	"github.com/Vanady39/cluer/platform/oidcauth"
)

type User struct {
	Subject   string `json:"subject"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatarUrl"`
}

type GetCurrentUserResponse struct {
	Data User `json:"data"`
}

func NewGetCurrentUserResponse(claims oidcauth.Claims) GetCurrentUserResponse {
	return GetCurrentUserResponse{
		Data: User{
			Subject:   claims.Subject,
			Email:     claims.Email,
			Name:      claims.Name,
			Username:  claims.PreferredUsername,
			AvatarURL: claims.Picture,
		},
	}
}
