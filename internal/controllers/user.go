package controllers

import (
	"net/http"

	"github.com/Vanady39/cluer/internal/domains"
	"github.com/Vanady39/cluer/internal/models/response"
	"github.com/gin-gonic/gin"
)

type (
	UserController struct {
		domain domains.UserDomainInterface
	}

	UserControllerInterface interface {
		GetCurrentUser(ctx *gin.Context)
	}
)

func NewUserController(domain domains.UserDomainInterface) *UserController {
	return &UserController{domain: domain}
}

// GetCurrentUser godoc
//
//	@Summary		Get the current user
//	@Description	Returns the currently authenticated user
//	@Tags			Users
//	@Produce		json
//	@Success		200	{object}	response.GetCurrentUserResponse
//	@Failure		500	{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/users/me [get]
func (uc *UserController) GetCurrentUser(ctx *gin.Context) {
	user, err := uc.domain.GetCurrentUser(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, response.NewGetCurrentUserResponse(user))
}
