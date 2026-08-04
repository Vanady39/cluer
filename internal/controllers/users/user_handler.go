package users

import (
	"net/http"

	"github.com/gin-gonic/gin"

	domainusers "github.com/Vanady39/cluer/internal/domains/users"
	apiresponse "github.com/Vanady39/cluer/internal/models/response"
	userresponse "github.com/Vanady39/cluer/internal/models/response/users"
)

type UserHandler struct {
	service domainusers.UserService
}

func NewUserHandler(service domainusers.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	user, err := h.service.GetCurrentUser(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, apiresponse.ErrorResponse{
			Error: apiresponse.Error{
				Code:    "INTERNAL_ERROR",
				Message: "Не удалось получить пользователя",
			},
		})

		return
	}

	c.JSON(http.StatusOK, userresponse.NewGetCurrentUserResponse(user))
}
