package users

import (
	"encoding/json"
	"net/http"

	response "github.com/Vanady39/cluer/internal/models/response/users"
	repo "github.com/Vanady39/cluer/internal/repositories/users"
)

func GetCurrentUser(w http.ResponseWriter, _ *http.Request) {
	user := repo.GetMockCurrentUser()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(response.GetUserResponse{
		Data: user,
	})
}
