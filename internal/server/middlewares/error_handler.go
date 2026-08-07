package middlewares

import (
	"net/http"

	"github.com/Vanady39/cluer/internal/controllers"
	"github.com/Vanady39/cluer/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// ErrorHandler turns whatever a handler pushed onto the context into the single
// error shape the API promises.
//
// The full error is logged, but only the curated message goes out: driver text
// and SQL fragments describe the schema to whoever asked, and an admin cannot
// act on them anyway.
func ErrorHandler(logger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		last := c.Errors.Last().Err
		logger.Trace().Int("err_count", len(c.Errors)).Msg("One or more errors occurred")

		base, ok := last.(controllers.BaseError)
		if !ok {
			logger.Error().Err(last).Msg("Error handler got an error of unknown type")
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorEnvelope{
				Error: models.ErrorBody{
					Code:    models.CodeInternal,
					Message: "An unknown error has occurred",
				},
			})
			return
		}

		status := base.StatusCode()
		if status >= http.StatusInternalServerError {
			logger.Error().Err(last).Int("status", status).Msg("Request failed")
		} else {
			logger.Debug().Err(last).Int("status", status).Msg("Request rejected")
		}

		body := models.ErrorBody{
			Code:    models.CodeForStatus(status),
			Message: base.Message(),
		}
		if coded, ok := last.(controllers.CodedError); ok {
			body.Code = coded.Code()
		}
		if detailed, ok := last.(controllers.DetailedError); ok {
			body.Details = detailed.ErrorDetails()
		}

		c.AbortWithStatusJSON(status, models.ErrorEnvelope{Error: body})
	}
}
