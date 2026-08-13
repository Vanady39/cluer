package errs_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/Vanady39/cluer/platform/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBindingError фиксирует обе ветки BindingError. Detail заполняется только
// в листингах, а сам тип разделяют tour/hint/runtime/params, поэтому легаси-ветка
// закреплена отдельно: правка формата ошибок листингов не должна тихо поменять
// тела 400 у остальных контроллеров.
func TestBindingError(t *testing.T) {
	t.Run("с Detail: error остаётся generic, деталь уходит в message", func(t *testing.T) {
		inner := errors.New("limit must be an integer between 1 and 50")
		err := &errs.BindingError{
			Err:    inner,
			Zone:   errs.Query,
			Code:   http.StatusBadRequest,
			Detail: "limit must be an integer between 1 and 50",
		}

		assert.Equal(t, "failed to bind query", err.Error())
		assert.Equal(t, "Failed to parse: limit must be an integer between 1 and 50", err.Message())
		assert.Equal(t, http.StatusBadRequest, err.StatusCode())

		payload := err.ToHTTPError()
		require.NotNil(t, payload)
		assert.Equal(t, "failed to bind query", payload.Error)
		assert.Equal(t, "Failed to parse: limit must be an integer between 1 and 50", payload.Message)

		assert.True(t, errors.Is(err, inner), "Unwrap обязан сохранять исходную ошибку для логов")
	})

	t.Run("без Detail: сохраняется прежняя форма для остальных контроллеров", func(t *testing.T) {
		inner := errors.New("unexpected end of JSON input")
		err := &errs.BindingError{
			Err:  inner,
			Zone: errs.Body,
			Code: http.StatusBadRequest,
		}

		assert.Equal(t, "failed to bind body: unexpected end of JSON input", err.Error())
		assert.Equal(t, "Failed to parse: invalid key/value in body", err.Message())

		payload := err.ToHTTPError()
		require.NotNil(t, payload)
		assert.Equal(t, "failed to bind body: unexpected end of JSON input", payload.Error)
		assert.Equal(t, "Failed to parse: invalid key/value in body", payload.Message)

		assert.True(t, errors.Is(err, inner))
	})

	t.Run("реализует BaseError, поэтому его подхватывает ErrorHandler", func(t *testing.T) {
		var _ errs.BaseError = &errs.BindingError{}
	})
}
