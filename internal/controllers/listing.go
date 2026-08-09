package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Vanady39/cluer/internal/domains"
	"github.com/Vanady39/cluer/internal/models/response"
	"github.com/gin-gonic/gin"
)

const (
	defaultListingsLimit = 20
	maxListingsLimit     = 50
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
//	@Summary		List and search listings
//	@Description	Returns listings, optionally filtering by title or description
//	@Tags			Listings
//	@Produce		json
//	@Param			q		query	string	false	"Case-insensitive text search"
//	@Param			limit	query	int		false	"Result limit (1-50, default 20)"
//	@Param			offset	query	int		false	"Result offset (default 0)"
//	@Success		200	{object}	response.GetListingsResponse
//	@Failure		400	{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Failure		500	{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/listings [get]
func (lc *ListingController) GetListings(ctx *gin.Context) {
	filter, err := listingFilter(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	listings, err := lc.domain.GetListings(ctx.Request.Context(), filter)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, response.NewGetListingsResponse(listings))
}

func listingFilter(ctx *gin.Context) (domains.ListingFilter, error) {
	limit, err := optionalListingQueryInt(ctx, "limit", defaultListingsLimit, 1, maxListingsLimit)
	if err != nil {
		return domains.ListingFilter{}, err
	}

	offset, err := optionalListingQueryInt(ctx, "offset", 0, 0, 0)
	if err != nil {
		return domains.ListingFilter{}, err
	}

	return domains.ListingFilter{
		Query:  strings.TrimSpace(ctx.Query("q")),
		Limit:  limit,
		Offset: offset,
	}, nil
}

func optionalListingQueryInt(ctx *gin.Context, name string, defaultValue, minValue, maxValue int) (int, error) {
	raw, present := ctx.GetQuery(name)
	if !present {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < minValue || (maxValue > 0 && value > maxValue) {
		message := fmt.Sprintf("%s must be an integer greater than or equal to %d", name, minValue)
		if maxValue > 0 {
			message = fmt.Sprintf("%s must be an integer between %d and %d", name, minValue, maxValue)
		}
		return 0, &BindingError{
			Err:  errors.New(message),
			Zone: Query,
			Code: http.StatusBadRequest,
		}
	}

	return value, nil
}
