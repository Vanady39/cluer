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
		List(ctx *gin.Context)
		Update(ctx *gin.Context)
		Delete(ctx *gin.Context)
		Reorder(ctx *gin.Context)
	}

	reorderRequest struct {
		HintIds []uuid.UUID `json:"hint_ids" binding:"required"`
	}
)

func NewHintController(domain domains.HintDomainInterface) *HintController {
	return &HintController{domain: domain}
}

// Create godoc
//
//	@Summary		Create a hint
//	@Description	Appends a hint to the tour's draft version
//	@Tags			Hints
//	@Accept			json
//	@Produce		json
//	@Param			tourId	path		string			true	"Tour ID"	format(uuid)
//	@Param			hint	body		domains.Hint	true	"Hint object"
//	@Success		201		{object}	domains.Hint
//	@Failure		400		{object}	models.ErrorEnvelope
//	@Failure		409		{object}	models.ErrorEnvelope
//	@Router			/tours/{tourId}/hints [post]
func (hc *HintController) Create(ctx *gin.Context) {
	tourId, err := pathUUID(ctx, "tourId")
	if err != nil {
		ctx.Error(err)
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
	ctx.JSON(http.StatusCreated, created)
}

// List godoc
//
//	@Summary		List draft hints
//	@Description	Returns the hints of the tour's draft version, ordered by step
//	@Tags			Hints
//	@Produce		json
//	@Param			tourId	path		string	true	"Tour ID"	format(uuid)
//	@Success		200		{array}		domains.Hint
//	@Failure		409		{object}	models.ErrorEnvelope
//	@Router			/tours/{tourId}/hints [get]
func (hc *HintController) List(ctx *gin.Context) {
	tourId, err := pathUUID(ctx, "tourId")
	if err != nil {
		ctx.Error(err)
		return
	}

	hints, err := hc.domain.ListByTour(ctx.Request.Context(), tourId)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, hints)
}

// Update godoc
//
//	@Summary		Update a hint
//	@Description	Edits a hint of the draft version; hints of published versions are immutable
//	@Tags			Hints
//	@Accept			json
//	@Produce		json
//	@Param			tourId	path		string			true	"Tour ID"	format(uuid)
//	@Param			hintId	path		string			true	"Hint ID"	format(uuid)
//	@Param			hint	body		domains.Hint	true	"Hint object"
//	@Success		200		{object}	domains.Hint
//	@Failure		409		{object}	models.ErrorEnvelope
//	@Router			/tours/{tourId}/hints/{hintId} [patch]
func (hc *HintController) Update(ctx *gin.Context) {
	tourId, err := pathUUID(ctx, "tourId")
	if err != nil {
		ctx.Error(err)
		return
	}
	hintId, err := pathUUID(ctx, "hintId")
	if err != nil {
		ctx.Error(err)
		return
	}

	hint := new(domains.Hint)
	if err := ctx.ShouldBindJSON(hint); err != nil {
		ctx.Error(&BindingError{Err: err, Zone: Body, Code: http.StatusBadRequest})
		return
	}

	updated, err := hc.domain.Update(ctx.Request.Context(), tourId, hintId, hint)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

// Delete godoc
//
//	@Summary		Delete a hint
//	@Description	Removes a hint from the draft and closes the gap in step numbering
//	@Tags			Hints
//	@Param			tourId	path	string	true	"Tour ID"	format(uuid)
//	@Param			hintId	path	string	true	"Hint ID"	format(uuid)
//	@Success		204		"No Content"
//	@Failure		404		{object}	models.ErrorEnvelope
//	@Router			/tours/{tourId}/hints/{hintId} [delete]
func (hc *HintController) Delete(ctx *gin.Context) {
	tourId, err := pathUUID(ctx, "tourId")
	if err != nil {
		ctx.Error(err)
		return
	}
	hintId, err := pathUUID(ctx, "hintId")
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := hc.domain.Delete(ctx.Request.Context(), tourId, hintId); err != nil {
		ctx.Error(err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// Reorder godoc
//
//	@Summary		Reorder hints
//	@Description	Takes the full list of hint ids in their new order
//	@Tags			Hints
//	@Accept			json
//	@Produce		json
//	@Param			tourId	path		string						true	"Tour ID"	format(uuid)
//	@Param			body	body		controllers.reorderRequest	true	"Ordered hint ids"
//	@Success		200		{array}		domains.Hint
//	@Failure		400		{object}	models.ErrorEnvelope
//	@Router			/tours/{tourId}/hints/order [put]
func (hc *HintController) Reorder(ctx *gin.Context) {
	tourId, err := pathUUID(ctx, "tourId")
	if err != nil {
		ctx.Error(err)
		return
	}

	req := new(reorderRequest)
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.Error(&BindingError{Err: err, Zone: Body, Code: http.StatusBadRequest})
		return
	}

	hints, err := hc.domain.Reorder(ctx.Request.Context(), tourId, req.HintIds)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, hints)
}
