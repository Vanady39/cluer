package response

import domain "github.com/Vanady39/cluer/internal/domains/users"

type GetUserResponse struct {
	Data domain.User `json:"data"`
}
