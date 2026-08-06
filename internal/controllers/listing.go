package controllers

import (
	"net/http"

	"github.com/Vanady39/cluer/internal/domains"
	"github.com/Vanady39/cluer/internal/models/response"
	"github.com/gin-gonic/gin"
)

type (
	ListingController struct {
		domain domains.ListingDomainInterface
	}

	ListingControllerInterface interface {
		GetListings(ctx *gin.Context)
	}
)

func NewListingController(domain domains.ListingDomainInterface) *ListingController {
	return &ListingController{domain: domain}
}

// GetListings godoc
//
//	@Summary		Get all listings
//	@Description	Returns every available listing
//	@Tags			Listings
//	@Produce		json
//	@Success		200	{object}	response.GetListingsResponse
//	@Failure		500	{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/listings [get]
func (lc *ListingController) GetListings(ctx *gin.Context) {
	listings, err := lc.domain.GetListings(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, response.NewGetListingsResponse(listings))
}
