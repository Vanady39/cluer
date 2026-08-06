package controllers

import (
	"net/http"

	"github.com/Vanady39/cluer/internal/domains"
	"github.com/gin-gonic/gin"
)

type (
	TourController struct {
		domain domains.TourDomainInterface
	}

	TourControllerInterface interface {
		Create(ctx *gin.Context)
		GetPublished(ctx *gin.Context)
	}
)

func NewTourController(domain domains.TourDomainInterface) *TourController {
	return &TourController{domain: domain}
}

// Create godoc
//
//	@Summary		Create a new tour
//	@Description	Creates a new onboarding tour draft
//	@Tags			Tours
//	@Accept			json
//	@Produce		json
//	@Param			tour	body	domains.Tour	true	"Tour object"
//	@Success		201		"Created"
//	@Failure		400		{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Failure		500		{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/tours [post]
func (tc *TourController) Create(ctx *gin.Context) {
	tour := new(domains.Tour)
	if err := ctx.ShouldBindJSON(tour); err != nil {
		ctx.Error(&BindingError{Err: err, Zone: Body, Code: http.StatusBadRequest})
		return
	}

	created, err := tc.domain.Create(ctx.Request.Context(), tour)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Header("Location", "/v1/tours/"+created.Id.String())
	ctx.Status(http.StatusCreated)
}

// GetPublished godoc
//
//	@Summary		Get published tours by path
//	@Description	Returns published tours with their hints for a specific target path
//	@Tags			Tours
//	@Produce		json
//	@Param			path	query		string	true	"Target path"	example(/dashboard)
//	@Success		200		{array}		domains.Tour
//	@Failure		400		{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Failure		500		{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/tours/published [get]
func (tc *TourController) GetPublished(ctx *gin.Context) {
	path := ctx.Query("path")
	if path == "" {
		ctx.Error(&BindingError{
			Err:  errMissingPath,
			Zone: Query,
			Code: http.StatusBadRequest,
		})
		return
	}

	tours, err := tc.domain.ListPublished(ctx.Request.Context(), path)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, tours)
}
