package controllers

import (
	"net/http"

	"github.com/Vanady39/cluer/internal/auth"
	"github.com/Vanady39/cluer/internal/models/response"
	"github.com/gin-gonic/gin"
)

type (
	UserController struct{}

	UserControllerInterface interface {
		GetCurrentUser(ctx *gin.Context)
	}

	MeController struct{}

	MeControllerInterface interface {
		GetCurrentAdmin(ctx *gin.Context)
	}
)

func NewUserController() *UserController {
	return &UserController{}
}

func NewMeController() *MeController {
	return &MeController{}
}

// GetCurrentUser godoc
//
//	@Summary		Get the current user
//	@Description	Returns the identity behind the presented id token: subject and profile fields. Any authenticated visitor qualifies, no group is required.
//	@Tags			Users
//	@Produce		json
//	@Security		Bearer
//	@Success		200	{object}	response.GetCurrentUserResponse
//	@Failure		401	{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/users/me [get]
func (uc *UserController) GetCurrentUser(ctx *gin.Context) {
	// Unreachable while the route sits behind OIDCAuth. Handled anyway so that
	// moving the route out of the authenticated chain fails loudly instead of
	// serving an empty identity — the shape a mocked user would have.
	claims, err := auth.FromGin(ctx)
	if err != nil {
		ctx.Error(&PermissionError{Err: err, Code: http.StatusUnauthorized})
		return
	}

	ctx.JSON(http.StatusOK, response.NewGetCurrentUserResponse(claims))
}

// GetCurrentAdmin godoc
//
//	@Summary		Get the current administrator
//	@Description	Returns the identity behind the presented id token: subject, profile fields and IdP groups.
//	@Tags			Users
//	@Produce		json
//	@Security		Bearer
//	@Success		200	{object}	response.GetCurrentAdminResponse
//	@Failure		401	{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Failure		403	{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/users/me [get]
func (mc *MeController) GetCurrentAdmin(ctx *gin.Context) {
	// Unreachable while the route sits behind OIDCAuth. Handled anyway so that
	// moving the route somewhere unauthenticated fails loudly instead of
	// serving an empty identity.
	claims, err := auth.FromGin(ctx)
	if err != nil {
		ctx.Error(&PermissionError{Err: err, Code: http.StatusUnauthorized})
		return
	}

	ctx.JSON(http.StatusOK, response.NewGetCurrentAdminResponse(claims))
}
