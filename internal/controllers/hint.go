package controllers

import (
	"net/http"

	"github.com/Vanady39/cluer/internal/domains"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type (
	HintController struct {
		domain domains.HintDomainInterface
	}

	HintControllerInterface interface {
		Create(ctx *gin.Context)
	}
)

func NewHintController(domain domains.HintDomainInterface) *HintController {
	return &HintController{domain: domain}
}

// Create godoc
//
//	@Summary		Create a hint for a tour
//	@Description	Creates a new hint for a specific tour
//	@Tags			Hints
//	@Accept			json
//	@Produce		json
//	@Param			tourId	path	string			true	"Tour ID"	format(uuid)
//	@Param			hint	body	domains.Hint	true	"Hint object"
//	@Success		201		"Created"
//	@Failure		400		{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Failure		404		{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Failure		409		{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Failure		500		{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/tours/{tourId}/hints [post]
func (hc *HintController) Create(ctx *gin.Context) {
	tourId, err := uuid.Parse(ctx.Param("tourId"))
	if err != nil {
		ctx.Error(&BindingError{Err: err, Zone: URI, Code: http.StatusBadRequest})
		return
	}

	hint := new(domains.Hint)
	if err := ctx.ShouldBindJSON(hint); err != nil {
		ctx.Error(&BindingError{Err: err, Zone: Body, Code: http.StatusBadRequest})
		return
	}

	created, err := hc.domain.Create(ctx.Request.Context(), tourId, hint)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Header("Location", "/v1/tours/"+tourId.String()+"/hints/"+created.Id.String())
	ctx.Status(http.StatusCreated)
}
