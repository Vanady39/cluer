package users

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	domainusers "github.com/Vanady39/cluer/internal/domains/users"
	apiresponse "github.com/Vanady39/cluer/internal/models/response"
	userresponse "github.com/Vanady39/cluer/internal/models/response/users"
)

type Service interface {
	GetCurrentUser(ctx context.Context) (domainusers.User, error)
}

type Handler struct {
	service Service
}

func New(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetCurrentUser(c *gin.Context) {
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
