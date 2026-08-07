package response

import "github.com/Vanady39/cluer/internal/domains"

type User struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
}

type GetCurrentUserResponse struct {
	Data User `json:"data"`
}

func NewGetCurrentUserResponse(user domains.User) GetCurrentUserResponse {
	return GetCurrentUserResponse{
		Data: User{
			ID:        user.ID,
			Name:      user.Name,
			AvatarURL: user.AvatarURL,
		},
	}
}
