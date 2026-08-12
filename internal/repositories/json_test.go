package repositories

import (
	"testing"

	"github.com/Vanady39/cluer/internal/domains"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalJSON(t *testing.T) {
	t.Run("nil возвращает nil", func(t *testing.T) {
		b, err := marshalJSON(nil)
		assert.NoError(t, err)
		assert.Nil(t, b)
	})

	t.Run("валидная структура", func(t *testing.T) {
		cfg := &domains.TriggerConfig{ScrollDepth: ptr(50)}
		b, err := marshalJSON(cfg)
		require.NoError(t, err)
		assert.Contains(t, string(b), `"scroll_depth":50`)
	})
}

func TestUnmarshalJSONB(t *testing.T) {
	t.Run("пустые данные", func(t *testing.T) {
		var cfg domains.TriggerConfig
		err := unmarshalJSONB(nil, &cfg)
		assert.NoError(t, err)
	})

	t.Run("валидный JSON", func(t *testing.T) {
		var cfg domains.TriggerConfig
		err := unmarshalJSONB([]byte(`{"scroll_depth":50}`), &cfg)
		require.NoError(t, err)
		assert.Equal(t, 50, *cfg.ScrollDepth)
	})

	t.Run("битый JSON возвращает ErrCorruptedData", func(t *testing.T) {
		var cfg domains.TriggerConfig
		err := unmarshalJSONB([]byte(`{invalid`), &cfg)
		require.Error(t, err)
		assert.ErrorIs(t, err, domains.ErrCorruptedData)
	})
}

func ptr[T any](v T) *T {
	return &v
}
