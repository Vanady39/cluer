package controllers

import (
	"net/http"
	"time"

	"github.com/Vanady39/cluer/internal/domains"
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
	}

	ResolveRequest struct {
		Url       string         `json:"url"`
		SubjectId string         `json:"subject_id" binding:"required"`
		SessionId string         `json:"session_id"`
		Props     map[string]any `json:"props"`
	}

	EventItem struct {
		EventKey      string            `json:"event_key" binding:"required"`
		Type          domains.EventType `json:"type" binding:"required"`
		TourId        uuid.UUID         `json:"tour_id"`
		TourVersionId uuid.UUID         `json:"tour_version_id"`
		HintId        *uuid.UUID        `json:"hint_id"`
		OccurredAt    *time.Time        `json:"occurred_at"`
		Payload       map[string]any    `json:"payload"`
	}

	EventBatchRequest struct {
		SubjectId string      `json:"subject_id" binding:"required"`
		SessionId string      `json:"session_id" binding:"required"`
		Events    []EventItem `json:"events" binding:"required"`
	}

	CreateAppRequest struct {
		Name           string   `json:"name" binding:"required"`
		AllowedOrigins []string `json:"allowed_origins"`
	}

	HealthResponse struct {
		Status string `json:"status" example:"ok"`
		DB     string `json:"db" example:"ok"`
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
//	@Param			X-App-Key	header		string							true	"Application public key"
//	@Param			body		body		controllers.ResolveRequest	true	"Resolve context"
//	@Success		200			{object}	domains.ResolveResult
//	@Success		204			"Nothing to show"
//	@Failure		401			{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/resolve [post]
func (rc *RuntimeController) Resolve(ctx *gin.Context) {
	appId, err := appIdFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	req := new(ResolveRequest)
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.Error(&BindingError{Err: err, Zone: Body, Code: http.StatusBadRequest})
		return
	}

	result, err := rc.domain.Resolve(ctx.Request.Context(), domains.ResolveRequest{
		AppId:     appId,
		Url:       req.Url,
		SubjectId: req.SubjectId,
		SessionId: req.SessionId,
		Props:     req.Props,
	})
	if err != nil {
		ctx.Error(err)
		return
	}

	if result == nil {
		ctx.Status(http.StatusNoContent)
		return
	}
	ctx.JSON(http.StatusOK, result)
}

// Ingest godoc
//
//	@Summary		Accept a batch of events
//	@Description	Idempotent on event_key, so a retry after a dropped connection cannot double the metrics
//	@Tags			Runtime
//	@Accept			json
//	@Produce		json
//	@Param			X-App-Key	header		string							true	"Application public key"
//	@Param			body		body		controllers.EventBatchRequest	true	"Event batch"
//	@Success		202			{object}	domains.EventBatchResult
//	@Failure		400			{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/events [post]
func (rc *RuntimeController) Ingest(ctx *gin.Context) {
	appId, err := appIdFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	req := new(EventBatchRequest)
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.Error(&BindingError{Err: err, Zone: Body, Code: http.StatusBadRequest})
		return
	}

	events := make([]domains.Event, 0, len(req.Events))
	for _, e := range req.Events {
		event := domains.Event{
			EventKey:      e.EventKey,
			TourId:        e.TourId,
			TourVersionId: e.TourVersionId,
			HintId:        e.HintId,
			Type:          e.Type,
			Payload:       e.Payload,
		}
		if e.OccurredAt != nil {
			event.OccurredAt = *e.OccurredAt
		}
		events = append(events, event)
	}

	result, err := rc.domain.Ingest(ctx.Request.Context(), appId, req.SubjectId, req.SessionId, events)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusAccepted, result)
}

// Health godoc
//
//	@Summary		Health check
//	@Description	Reports whether the process and its database are both usable
//	@Tags			Runtime
//	@Produce		json
//	@Success		200	{object}	controllers.HealthResponse
//	@Failure		503	{object}	controllers.HealthResponse
//	@Router			/health [get]
func (rc *RuntimeController) Health(ctx *gin.Context) {
	if err := rc.domain.Ping(ctx.Request.Context()); err != nil {
		ctx.JSON(http.StatusServiceUnavailable, HealthResponse{Status: "degraded", DB: "down"})
		return
	}
	ctx.JSON(http.StatusOK, HealthResponse{Status: "ok", DB: "ok"})
}

// Analytics godoc
//
//	@Summary		Tour funnel and totals
//	@Description	Per-hint funnel for one version over one period
//	@Tags			Analytics
//	@Produce		json
//	@Param			tourId		path		string	true	"Tour ID"						format(uuid)
//	@Param			versionId	query		string	false	"Version ID, defaults to published"	format(uuid)
//	@Param			from		query		string	false	"Period start, RFC3339"
//	@Param			to			query		string	false	"Period end, RFC3339"
//	@Success		200			{object}	domains.Analytics
//	@Failure		404			{object}	github_com_Vanady39_cluer_internal_models.HTTPError
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
	ctx.JSON(http.StatusOK, analytics)
}

// CreateApp godoc
//
//	@Summary		Register a consumer application
//	@Description	Issues the public key that goes into the SDK script tag
//	@Tags			Apps
//	@Accept			json
//	@Produce		json
//	@Param			body	body		controllers.CreateAppRequest	true	"Application"
//	@Success		201		{object}	domains.App
//	@Failure		400		{object}	github_com_Vanady39_cluer_internal_models.HTTPError
//	@Router			/apps [post]
func (rc *RuntimeController) CreateApp(ctx *gin.Context) {
	req := new(CreateAppRequest)
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.Error(&BindingError{Err: err, Zone: Body, Code: http.StatusBadRequest})
		return
	}

	created, err := rc.domain.CreateApp(ctx.Request.Context(), &domains.App{
		Name:           req.Name,
		AllowedOrigins: req.AllowedOrigins,
	})
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Header("Location", "/v1/apps/"+created.Id.String())
	ctx.JSON(http.StatusCreated, created)
}

// ListApps godoc
//
//	@Summary		List consumer applications
//	@Tags			Apps
//	@Produce		json
//	@Success		200	{array}	domains.App
//	@Router			/apps [get]
func (rc *RuntimeController) ListApps(ctx *gin.Context) {
	apps, err := rc.domain.ListApps(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, apps)
}

func appIdFromContext(ctx *gin.Context) (uuid.UUID, error) {
	value, ok := ctx.Get("app_id")
	if !ok {
		return uuid.Nil, &PermissionError{Err: errMissingAppKey, Code: http.StatusUnauthorized}
	}
	appId, ok := value.(uuid.UUID)
	if !ok {
		return uuid.Nil, &PermissionError{Err: errMissingAppKey, Code: http.StatusUnauthorized}
	}
	return appId, nil
}
