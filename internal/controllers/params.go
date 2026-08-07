package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// pathUUID reads a UUID from the route. A malformed id is a client mistake, not
// a missing entity, so it answers 400 rather than letting the repository return
// a confusing 404.
func pathUUID(ctx *gin.Context, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(ctx.Param(name))
	if err != nil {
		return uuid.Nil, &BindingError{Err: err, Zone: URI, Code: http.StatusBadRequest}
	}
	return id, nil
}

func requiredUUIDQuery(ctx *gin.Context, name string) (uuid.UUID, error) {
	raw := ctx.Query(name)
	if raw == "" {
		return uuid.Nil, &BindingError{Err: errMissingAppId, Zone: Query, Code: http.StatusBadRequest}
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, &BindingError{Err: err, Zone: Query, Code: http.StatusBadRequest}
	}
	return id, nil
}

// optionalUUIDQuery returns uuid.Nil when the parameter is absent, which callers
// read as "use the default".
func optionalUUIDQuery(ctx *gin.Context, name string) (uuid.UUID, error) {
	raw := ctx.Query(name)
	if raw == "" {
		return uuid.Nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, &BindingError{Err: err, Zone: Query, Code: http.StatusBadRequest}
	}
	return id, nil
}

// timeRange reads the reporting period, defaulting to the last 30 days.
//
// The period filters received_at rather than occurred_at: client clocks can be
// skewed by hours, and a report "for yesterday" built on them would quietly
// include or drop the wrong sessions.
func timeRange(ctx *gin.Context) (time.Time, time.Time, error) {
	to := time.Now().UTC()
	from := to.AddDate(0, 0, -30)

	if raw := ctx.Query("from"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return from, to, &BindingError{Err: errBadTimeFormat, Zone: Query, Code: http.StatusBadRequest}
		}
		from = parsed
	}
	if raw := ctx.Query("to"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return from, to, &BindingError{Err: errBadTimeFormat, Zone: Query, Code: http.StatusBadRequest}
		}
		to = parsed
	}
	return from, to, nil
}
