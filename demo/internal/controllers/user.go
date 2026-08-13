package controllers

import (
	"net/http"

	"github.com/Vanady39/cluer/demo/internal/http/response"
	"github.com/Vanady39/cluer/platform/errs"
	"github.com/Vanady39/cluer/platform/oidcauth"
	"github.com/gin-gonic/gin"
)

type (
	UserController struct{}

	UserControllerInterface interface {
		GetCurrentUser(ctx *gin.Context)
	}
)

func NewUserController() *UserController {
	return &UserController{}
}

// GetCurrentUser godoc
//
//	@Summary		Get the current user
//	@Description	Returns the identity behind the presented id token: subject and profile fields. Any authenticated visitor qualifies, no group is required.
//	@Tags			Users
//	@Produce		json
//	@Security		Bearer
//	@Success		200	{object}	response.GetCurrentUserResponse
//	@Failure		401	{object}	errs.HTTPError
//	@Router			/users/me [get]
func (uc *UserController) GetCurrentUser(ctx *gin.Context) {
	// Unreachable while the route sits behind OIDCAuth. Handled anyway so that
	// moving the route out of the authenticated chain fails loudly instead of
	// serving an empty identity — the shape a mocked user would have.
	claims, err := oidcauth.FromGin(ctx)
	if err != nil {
		ctx.Error(&errs.PermissionError{Err: err, Code: http.StatusUnauthorized})
		return
	}

	ctx.JSON(http.StatusOK, response.NewGetCurrentUserResponse(claims))
}
