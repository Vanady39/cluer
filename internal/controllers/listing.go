package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

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
//	@Summary		List and search listings
//	@Description	Returns listings, optionally filtering by title or description
//	@Tags			Listings
//	@Produce		json
//	@Param			q		query	string	false	"Case-insensitive text search (up to 200 characters)"
//	@Param			limit	query	int		false	"Result limit (1-50, default 20)"
//	@Param			offset	query	int		false	"Result offset (0-100000, default 0)"
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

	page, err := lc.domain.GetListings(ctx.Request.Context(), filter)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, response.NewGetListingsResponse(page))
}

func listingFilter(ctx *gin.Context) (domains.ListingFilter, error) {
	query := strings.TrimSpace(ctx.Query("q"))
	if utf8.RuneCountInString(query) > domains.MaxListingsQueryRunes {
		return domains.ListingFilter{}, invalidListingQuery("q must not exceed %d characters", domains.MaxListingsQueryRunes)
	}

	maxLimit := domains.MaxListingsLimit
	limit, err := optionalListingQueryInt(ctx, "limit", domains.DefaultListingsLimit, 1, &maxLimit)
	if err != nil {
		return domains.ListingFilter{}, err
	}

	maxOffset := domains.MaxListingsOffset
	offset, err := optionalListingQueryInt(ctx, "offset", 0, 0, &maxOffset)
	if err != nil {
		return domains.ListingFilter{}, err
	}

	return domains.ListingFilter{
		Query:  query,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func optionalListingQueryInt(ctx *gin.Context, name string, defaultValue, minValue int, maxValue *int) (int, error) {
	raw, present := ctx.GetQuery(name)
	if !present || raw == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < minValue || (maxValue != nil && value > *maxValue) {
		message := "%s must be an integer greater than or equal to %d"
		args := []any{name, minValue}
		if maxValue != nil {
			message = "%s must be an integer between %d and %d"
			args = []any{name, minValue, *maxValue}
		}
		return 0, invalidListingQuery(message, args...)
	}

	return value, nil
}

func invalidListingQuery(format string, args ...any) error {
	detail := fmt.Sprintf(format, args...)
	return &BindingError{
		Err:    errors.New(detail),
		Zone:   Query,
		Code:   http.StatusBadRequest,
		Detail: detail,
	}
}
