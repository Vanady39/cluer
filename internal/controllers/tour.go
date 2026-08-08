package controllers

import (
	"net/http"

	"github.com/Vanady39/cluer/internal/domains"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type (
	TourController struct {
		domain domains.TourDomainInterface
	}

	TourControllerInterface interface {
		Create(ctx *gin.Context)
		List(ctx *gin.Context)
		Card(ctx *gin.Context)
		UpdateMeta(ctx *gin.Context)
		Archive(ctx *gin.Context)
		GetPublished(ctx *gin.Context)

		ListVersions(ctx *gin.Context)
		GetVersion(ctx *gin.Context)
		CreateDraft(ctx *gin.Context)
		UpdateDraft(ctx *gin.Context)
		Publish(ctx *gin.Context)
		Rollback(ctx *gin.Context)
	}

	RollbackRequest struct {
		ToVersionId uuid.UUID `json:"to_version_id" binding:"required"`
	}
)

func NewTourController(domain domains.TourDomainInterface) *TourController {
	return &TourController{domain: domain}
}

// Create godoc
//
//	@Summary		Create a new tour
//	@Description	Creates a tour together with its first draft version
//	@Tags			Tours
//	@Accept			json
//	@Produce		json
//	@Param			appId	query		string			true	"App ID"	format(uuid)
//	@Param			tour	body		domains.Tour	true	"Tour object"
//	@Success		201		{object}	domains.Tour
//	@Failure		400		{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Failure		401		{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/tours [post]
func (tc *TourController) Create(ctx *gin.Context) {
	appId, err := requiredUUIDQuery(ctx, "appId")
	if err != nil {
		ctx.Error(err)
		return
	}

	tour := new(domains.Tour)
	if err := ctx.ShouldBindJSON(tour); err != nil {
		ctx.Error(&BindingError{Err: err, Zone: Body, Code: http.StatusBadRequest})
		return
	}

	created, err := tc.domain.Create(ctx.Request.Context(), appId, tour)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Header("Location", "/v1/tours/"+created.Id.String())
	ctx.JSON(http.StatusCreated, created)
}

// List godoc
//
//	@Summary		List tours
//	@Description	Returns every tour of an app with the status of its current version
//	@Tags			Tours
//	@Produce		json
//	@Param			appId	query		string	true	"App ID"	format(uuid)
//	@Success		200		{array}		domains.Tour
//	@Failure		400		{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/tours [get]
func (tc *TourController) List(ctx *gin.Context) {
	appId, err := requiredUUIDQuery(ctx, "appId")
	if err != nil {
		ctx.Error(err)
		return
	}

	tours, err := tc.domain.List(ctx.Request.Context(), appId)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, tours)
}

// Card godoc
//
//	@Summary		Get a tour card
//	@Description	Returns the tour with both its published and draft versions
//	@Tags			Tours
//	@Produce		json
//	@Param			tourId	path		string	true	"Tour ID"	format(uuid)
//	@Success		200		{object}	domains.TourCard
//	@Failure		404		{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/tours/{tourId} [get]
func (tc *TourController) Card(ctx *gin.Context) {
	tourId, err := pathUUID(ctx, "tourId")
	if err != nil {
		ctx.Error(err)
		return
	}

	card, err := tc.domain.Card(ctx.Request.Context(), tourId)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, card)
}

// UpdateMeta godoc
//
//	@Summary		Update tour metadata
//	@Description	Changes title, description, enabled or priority without creating a version
//	@Tags			Tours
//	@Accept			json
//	@Produce		json
//	@Param			tourId	path		string					true	"Tour ID"	format(uuid)
//	@Param			patch	body		domains.TourMetaPatch	true	"Fields to change"
//	@Success		200		{object}	domains.Tour
//	@Failure		404		{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/tours/{tourId} [patch]
func (tc *TourController) UpdateMeta(ctx *gin.Context) {
	tourId, err := pathUUID(ctx, "tourId")
	if err != nil {
		ctx.Error(err)
		return
	}

	patch := new(domains.TourMetaPatch)
	if err := ctx.ShouldBindJSON(patch); err != nil {
		ctx.Error(&BindingError{Err: err, Zone: Body, Code: http.StatusBadRequest})
		return
	}

	updated, err := tc.domain.UpdateMeta(ctx.Request.Context(), tourId, *patch)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

// Archive godoc
//
//	@Summary		Archive a tour
//	@Description	Soft delete: the tour stops resolving but its analytics history survives
//	@Tags			Tours
//	@Param			tourId	path	string	true	"Tour ID"	format(uuid)
//	@Success		204		"No Content"
//	@Failure		404		{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/tours/{tourId} [delete]
func (tc *TourController) Archive(ctx *gin.Context) {
	tourId, err := pathUUID(ctx, "tourId")
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := tc.domain.Archive(ctx.Request.Context(), tourId); err != nil {
		ctx.Error(err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// GetPublished godoc
//
//	@Summary		Get published tours by path
//	@Description	Returns published tours with their hints for a target path
//	@Tags			Tours
//	@Produce		json
//	@Param			appId	query		string	true	"App ID"		format(uuid)
//	@Param			path	query		string	true	"Target path"	example(/dashboard)
//	@Success		200		{array}		domains.Tour
//	@Failure		400		{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/tours/published [get]
func (tc *TourController) GetPublished(ctx *gin.Context) {
	appId, err := requiredUUIDQuery(ctx, "appId")
	if err != nil {
		ctx.Error(err)
		return
	}

	path := ctx.Query("path")
	if path == "" {
		ctx.Error(&BindingError{Err: errMissingPath, Zone: Query, Code: http.StatusBadRequest})
		return
	}

	tours, err := tc.domain.ListPublished(ctx.Request.Context(), appId, path)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, tours)
}

// ListVersions godoc
//
//	@Summary		List tour versions
//	@Description	Version history without hint bodies, newest first
//	@Tags			Versions
//	@Produce		json
//	@Param			tourId	path		string	true	"Tour ID"	format(uuid)
//	@Success		200		{array}		domains.TourVersion
//	@Router			/tours/{tourId}/versions [get]
func (tc *TourController) ListVersions(ctx *gin.Context) {
	tourId, err := pathUUID(ctx, "tourId")
	if err != nil {
		ctx.Error(err)
		return
	}

	versions, err := tc.domain.ListVersions(ctx.Request.Context(), tourId)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, versions)
}

// GetVersion godoc
//
//	@Summary		Get one version
//	@Description	Returns a version with its hints, for diffing and preview
//	@Tags			Versions
//	@Produce		json
//	@Param			versionId	path		string	true	"Version ID"	format(uuid)
//	@Success		200			{object}	domains.TourVersion
//	@Failure		404			{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/versions/{versionId} [get]
func (tc *TourController) GetVersion(ctx *gin.Context) {
	versionId, err := pathUUID(ctx, "versionId")
	if err != nil {
		ctx.Error(err)
		return
	}

	version, err := tc.domain.GetVersion(ctx.Request.Context(), versionId)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, version)
}

// CreateDraft godoc
//
//	@Summary		Fork a draft from the published version
//	@Description	Entry point to editing: the published version itself is never modified
//	@Tags			Versions
//	@Produce		json
//	@Param			tourId	path		string	true	"Tour ID"	format(uuid)
//	@Success		201		{object}	domains.TourVersion
//	@Failure		409		{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/tours/{tourId}/draft [post]
func (tc *TourController) CreateDraft(ctx *gin.Context) {
	tourId, err := pathUUID(ctx, "tourId")
	if err != nil {
		ctx.Error(err)
		return
	}

	draft, err := tc.domain.CreateDraft(ctx.Request.Context(), tourId)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Header("Location", "/v1/versions/"+draft.Id.String())
	ctx.JSON(http.StatusCreated, draft)
}

// UpdateDraft godoc
//
//	@Summary		Update the draft content
//	@Description	Saves work in progress; validation is enforced at publish time, not here
//	@Tags			Versions
//	@Accept			json
//	@Produce		json
//	@Param			tourId	path		string				true	"Tour ID"	format(uuid)
//	@Param			patch	body		domains.DraftPatch	true	"Fields to change"
//	@Success		200		{object}	domains.TourVersion
//	@Failure		409		{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/tours/{tourId}/draft [patch]
func (tc *TourController) UpdateDraft(ctx *gin.Context) {
	tourId, err := pathUUID(ctx, "tourId")
	if err != nil {
		ctx.Error(err)
		return
	}

	patch := new(domains.DraftPatch)
	if err := ctx.ShouldBindJSON(patch); err != nil {
		ctx.Error(&BindingError{Err: err, Zone: Body, Code: http.StatusBadRequest})
		return
	}

	updated, err := tc.domain.UpdateDraft(ctx.Request.Context(), tourId, *patch)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

// Publish godoc
//
//	@Summary		Publish the draft
//	@Description	Archives the live version and promotes the draft, atomically
//	@Tags			Versions
//	@Produce		json
//	@Param			tourId	path		string	true	"Tour ID"	format(uuid)
//	@Success		200		{object}	domains.TourVersion
//	@Failure		409		{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Failure		422		{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/tours/{tourId}/publish [post]
func (tc *TourController) Publish(ctx *gin.Context) {
	tourId, err := pathUUID(ctx, "tourId")
	if err != nil {
		ctx.Error(err)
		return
	}

	published, err := tc.domain.Publish(ctx.Request.Context(), tourId)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, published)
}

// Rollback godoc
//
//	@Summary		Roll back to an earlier version
//	@Description	Copies the chosen version into a new one and publishes it; history is appended, never rewritten
//	@Tags			Versions
//	@Accept			json
//	@Produce		json
//	@Param			tourId	path		string					true	"Tour ID"	format(uuid)
//	@Param			body	body		controllers.RollbackRequest	true	"Target version"
//	@Success		200		{object}	domains.TourVersion
//	@Failure		409		{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/tours/{tourId}/rollback [post]
func (tc *TourController) Rollback(ctx *gin.Context) {
	tourId, err := pathUUID(ctx, "tourId")
	if err != nil {
		ctx.Error(err)
		return
	}

	req := new(RollbackRequest)
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.Error(&BindingError{Err: err, Zone: Body, Code: http.StatusBadRequest})
		return
	}

	published, err := tc.domain.Rollback(ctx.Request.Context(), tourId, req.ToVersionId)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, published)
}
