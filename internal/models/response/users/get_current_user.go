package users

import domainusers "github.com/Vanady39/cluer/internal/domains/users"

type User struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
}

type GetCurrentUserResponse struct {
	Data User `json:"data"`
}

func NewGetCurrentUserResponse(user domainusers.User) GetCurrentUserResponse {
	return GetCurrentUserResponse{
		Data: User{
			ID:        user.ID,
			Name:      user.Name,
			AvatarURL: user.AvatarURL,
		},
	}
}
