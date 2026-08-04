package users

import domains "github.com/Vanady39/cluer/internal/domains/users"

func GetMockCurrentUser() domains.User {
	return domains.User{
		ID:        1,
		Name:      "Иванов Иван",
		AvatarURL: "https://example.com/avatar.png",
	}
}
