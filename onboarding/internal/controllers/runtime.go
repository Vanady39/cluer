package controllers

import (
	"net/http"

	"github.com/Vanady39/cluer/onboarding/internal/domains"
	"github.com/Vanady39/cluer/onboarding/internal/http/request"
	"github.com/Vanady39/cluer/onboarding/internal/http/response"
	"github.com/Vanady39/cluer/platform/errs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type (
	RuntimeController struct {
		domain domains.RuntimeDomainInterface
	}

	RuntimeControllerInterface interface {
		Resolve(ctx *gin.Context)
		Ingest(ctx *gin.Context)
		Health(ctx *gin.Context)
		Analytics(ctx *gin.Context)
		CreateApp(ctx *gin.Context)
		ListApps(ctx *gin.Context)
		UpdateApp(ctx *gin.Context)
	}
)

func NewRuntimeController(domain domains.RuntimeDomainInterface) *RuntimeController {
	return &RuntimeController{domain: domain}
}

// Resolve godoc
//
//	@Summary		Resolve a tour for a subject
//	@Description	The single request the SDK makes on init. Returns at most one tour, whole.
//	@Tags			Runtime
//	@Accept			json
//	@Produce		json
//	@Param			X-App-Key	header		string					true	"Application public key"
//	@Param			body		body		request.ResolveRequest			true	"Resolve context"
//	@Success		200			{object}	response.ResolveResult
//	@Success		204			"Nothing to show"
//	@Failure		401			{object}	errs.HTTPError
//	@Router			/resolve [post]
func (rc *RuntimeController) Resolve(ctx *gin.Context) {
	appId, err := appIdFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	body := new(request.ResolveRequest)
	if err := ctx.ShouldBindJSON(body); err != nil {
		ctx.Error(&errs.BindingError{Err: err, Zone: errs.Body, Code: http.StatusBadRequest})
		return
	}

	result, err := rc.domain.Resolve(ctx.Request.Context(), body.ToDomain(appId))
	if err != nil {
		ctx.Error(err)
		return
	}

	if result == nil {
		ctx.Status(http.StatusNoContent)
		return
	}
	ctx.JSON(http.StatusOK, response.NewResolveResult(result))
}

// UpdateApp godoc
// @Summary		Update an application
// @Tags		Apps
// @Accept		json
// @Produce	json
// @Param		appId	path		string						true	"App ID"	format(uuid)
// @Param		body	body		request.UpdateAppRequest	true	"Fields to update"
// @Success	200		{object}	response.App
// @Failure	400		{object}	errs.HTTPError
// @Failure	404		{object}	errs.HTTPError
// @Router		/apps/{appId} [patch]
func (rc *RuntimeController) UpdateApp(ctx *gin.Context) {
	appId, err := pathUUID(ctx, "appId")
	if err != nil {
		ctx.Error(err)
		return
	}

	body := new(request.UpdateAppRequest)
	if err := ctx.ShouldBindJSON(body); err != nil {
		ctx.Error(&errs.BindingError{Err: err, Zone: errs.Body, Code: http.StatusBadRequest})
		return
	}

	updated, err := rc.domain.UpdateApp(ctx.Request.Context(), appId, body.Name, body.AllowedOrigins)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response.NewApp(updated))
}

// Ingest godoc
//
//	@Summary		Accept a batch of events
//	@Description	Idempotent on event_key, so a retry after a dropped connection cannot double the metrics
//	@Tags			Runtime
//	@Accept			json
//	@Produce		json
//	@Param			X-App-Key	header		string						true	"Application public key"
//	@Param			body		body		request.EventBatchRequest			true	"Event batch"
//	@Success		202			{object}	response.EventBatchResult
//	@Failure		400			{object}	errs.HTTPError
//	@Router			/events [post]
func (rc *RuntimeController) Ingest(ctx *gin.Context) {
	appId, err := appIdFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	body := new(request.EventBatchRequest)
	if err := ctx.ShouldBindJSON(body); err != nil {
		ctx.Error(&errs.BindingError{Err: err, Zone: errs.Body, Code: http.StatusBadRequest})
		return
	}

	result, err := rc.domain.Ingest(
		ctx.Request.Context(), appId, body.SubjectId, body.SessionId, body.EventsToDomain(),
	)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusAccepted, response.NewEventBatchResult(result))
}

// Health godoc
//
//	@Summary		Health check
//	@Description	Reports whether the process and its database are both usable
//	@Tags			Runtime
//	@Produce		json
//	@Success		200	{object}	response.Health
//	@Failure		503	{object}	response.Health
//	@Router			/health [get]
func (rc *RuntimeController) Health(ctx *gin.Context) {
	if err := rc.domain.Ping(ctx.Request.Context()); err != nil {
		ctx.JSON(http.StatusServiceUnavailable, response.Health{Status: "degraded", DB: "down"})
		return
	}
	ctx.JSON(http.StatusOK, response.Health{Status: "ok", DB: "ok"})
}

// Analytics godoc
//
//	@Summary		Tour funnel and totals
//	@Description	Per-hint funnel for one version over one period
//	@Tags			Analytics
//	@Produce		json
//	@Param			tourId		path		string	true	"Tour ID"							format(uuid)
//	@Param			versionId	query		string	false	"Version ID, defaults to published"	format(uuid)
//	@Param			from		query		string	false	"Period start, RFC3339"
//	@Param			to			query		string	false	"Period end, RFC3339"
//	@Success		200			{object}	response.Analytics
//	@Failure		404			{object}	errs.HTTPError
//	@Router			/tours/{tourId}/analytics [get]
func (rc *RuntimeController) Analytics(ctx *gin.Context) {
	tourId, err := pathUUID(ctx, "tourId")
	if err != nil {
		ctx.Error(err)
		return
	}
	versionId, err := optionalUUIDQuery(ctx, "versionId")
	if err != nil {
		ctx.Error(err)
		return
	}
	from, to, err := timeRange(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	analytics, err := rc.domain.Analytics(ctx.Request.Context(), tourId, versionId, from, to)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response.NewAnalytics(analytics))
}

// CreateApp godoc
//
//	@Summary		Register a consumer application
//	@Description	Issues the public key that goes into the SDK script tag
//	@Tags			Apps
//	@Accept			json
//	@Produce		json
//	@Param			body	body		request.CreateAppRequest	true	"Application"
//	@Success		201		{object}	response.App
//	@Failure		400		{object}	errs.HTTPError
//	@Router			/apps [post]
func (rc *RuntimeController) CreateApp(ctx *gin.Context) {
	body := new(request.CreateAppRequest)
	if err := ctx.ShouldBindJSON(body); err != nil {
		ctx.Error(&errs.BindingError{Err: err, Zone: errs.Body, Code: http.StatusBadRequest})
		return
	}

	created, err := rc.domain.CreateApp(ctx.Request.Context(), body.ToDomain())
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Header("Location", "/v1/apps/"+created.Id.String())
	ctx.JSON(http.StatusCreated, response.NewApp(created))
}

// ListApps godoc
//
//	@Summary		List consumer applications
//	@Tags			Apps
//	@Produce		json
//	@Success		200	{array}	response.App
//	@Router			/apps [get]
func (rc *RuntimeController) ListApps(ctx *gin.Context) {
	apps, err := rc.domain.ListApps(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response.NewApps(apps))
}

func appIdFromContext(ctx *gin.Context) (uuid.UUID, error) {
	value, ok := ctx.Get("app_id")
	if !ok {
		return uuid.Nil, &errs.PermissionError{Err: errMissingAppKey, Code: http.StatusUnauthorized}
	}
	appId, ok := value.(uuid.UUID)
	if !ok {
		return uuid.Nil, &errs.PermissionError{Err: errMissingAppKey, Code: http.StatusUnauthorized}
	}
	return appId, nil
}
