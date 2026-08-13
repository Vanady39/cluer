package controllers

import (
	"net/http"

	"github.com/Vanady39/cluer/onboarding/internal/http/response"
	"github.com/Vanady39/cluer/platform/errs"
	"github.com/Vanady39/cluer/platform/oidcauth"
	"github.com/gin-gonic/gin"
)

type (
	MeController struct{}

	MeControllerInterface interface {
		GetCurrentAdmin(ctx *gin.Context)
	}
)

func NewMeController() *MeController {
	return &MeController{}
}

// GetCurrentAdmin godoc
//
//	@Summary		Get the current administrator
//	@Description	Returns the identity behind the presented id token: subject, profile fields and IdP groups.
//	@Tags			Users
//	@Produce		json
//	@Security		Bearer
//	@Success		200	{object}	response.GetCurrentAdminResponse
//	@Failure		401	{object}	errs.HTTPError
//	@Failure		403	{object}	errs.HTTPError
//	@Router			/users/me [get]
func (mc *MeController) GetCurrentAdmin(ctx *gin.Context) {
	// Unreachable while the route sits behind OIDCAuth. Handled anyway so that
	// moving the route somewhere unauthenticated fails loudly instead of
	// serving an empty identity.
	claims, err := oidcauth.FromGin(ctx)
	if err != nil {
		ctx.Error(&errs.PermissionError{Err: err, Code: http.StatusUnauthorized})
		return
	}

	ctx.JSON(http.StatusOK, response.NewGetCurrentAdminResponse(claims))
}
