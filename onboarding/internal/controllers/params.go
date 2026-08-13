package controllers

import (
	"github.com/Vanady39/cluer/platform/errs"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func pathUUID(ctx *gin.Context, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(ctx.Param(name))
	if err != nil {
		return uuid.Nil, &errs.BindingError{Err: err, Zone: errs.URI, Code: http.StatusBadRequest}
	}
	return id, nil
}

func requiredUUIDQuery(ctx *gin.Context, name string) (uuid.UUID, error) {
	raw := ctx.Query(name)
	if raw == "" {
		return uuid.Nil, &errs.BindingError{Err: errMissingAppId, Zone: errs.Query, Code: http.StatusBadRequest}
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, &errs.BindingError{Err: err, Zone: errs.Query, Code: http.StatusBadRequest}
	}
	return id, nil
}

func optionalUUIDQuery(ctx *gin.Context, name string) (uuid.UUID, error) {
	raw := ctx.Query(name)
	if raw == "" {
		return uuid.Nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, &errs.BindingError{Err: err, Zone: errs.Query, Code: http.StatusBadRequest}
	}
	return id, nil
}

func timeRange(ctx *gin.Context) (time.Time, time.Time, error) {
	to := time.Now().UTC()
	from := to.AddDate(0, 0, -30)

	if raw := ctx.Query("from"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return from, to, &errs.BindingError{Err: errBadTimeFormat, Zone: errs.Query, Code: http.StatusBadRequest}
		}
		from = parsed
	}
	if raw := ctx.Query("to"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return from, to, &errs.BindingError{Err: errBadTimeFormat, Zone: errs.Query, Code: http.StatusBadRequest}
		}
		to = parsed
	}
	return from, to, nil
}
