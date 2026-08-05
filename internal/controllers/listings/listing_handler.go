package listings

import (
	"net/http"

	"github.com/gin-gonic/gin"

	domainlistings "github.com/Vanady39/cluer/internal/domains/listings"
	apiresponse "github.com/Vanady39/cluer/internal/models/response"
	listingresponse "github.com/Vanady39/cluer/internal/models/response/listings"
)

type ListingHandler struct {
	service domainlistings.ListingService
}

func NewListingHandler(
	service domainlistings.ListingService,
) *ListingHandler {
	return &ListingHandler{service: service}
}

func (h *ListingHandler) GetListings(c *gin.Context) {
	listings, err := h.service.GetListings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, apiresponse.ErrorResponse{
			Error: apiresponse.Error{
				Code:    "INTERNAL_ERROR",
				Message: "Не удалось получить объявления",
			},
		})
		return
	}

	c.JSON(
		http.StatusOK,
		listingresponse.NewGetListingsResponse(listings),
	)
}