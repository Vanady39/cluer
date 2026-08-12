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

	t.Run("typed nil *TriggerConfig возвращает nil (не JSONB null)", func(t *testing.T) {
		var cfg *domains.TriggerConfig
		b, err := marshalJSON(cfg)
		require.NoError(t, err)
		assert.Nil(t, b, "typed nil должен давать SQL NULL, а не JSONB null")
	})

	t.Run("валидная структура", func(t *testing.T) {
		cfg := &domains.TriggerConfig{ScrollDepth: ptr(50)}
		b, err := marshalJSON(cfg)
		require.NoError(t, err)
		assert.Contains(t, string(b), `"scroll_depth":50`)
	})
}

func TestUnmarshalJSONB(t *testing.T) {
	t.Run("литеральный nil возвращает nil", func(t *testing.T) {
		b, err := marshalJSON(nil)
		assert.NoError(t, err)
		assert.Nil(t, b)
	})

	t.Run("typed nil *TriggerConfig возвращает nil", func(t *testing.T) {
		var cfg *domains.TriggerConfig
		b, err := marshalJSON(cfg)
		require.NoError(t, err)
		assert.Nil(t, b, "typed nil должен давать SQL NULL, а не JSONB null")
	})

	t.Run("пустые данные", func(t *testing.T) {
		var cfg domains.TriggerConfig
		err := unmarshalJSONB(nil, &cfg)
		assert.NoError(t, err)
	})
	t.Run("typed nil *Audience возвращает nil", func(t *testing.T) {
		var aud *domains.Audience
		b, err := marshalJSON(aud)
		require.NoError(t, err)
		assert.Nil(t, b)
	})

	t.Run("typed nil []AudienceRule возвращает nil", func(t *testing.T) {
		var rules []domains.AudienceRule
		b, err := marshalJSON(rules)
		require.NoError(t, err)
		assert.Nil(t, b)
	})
	t.Run("пустой (не nil) слайс даёт []", func(t *testing.T) {
		rules := []domains.AudienceRule{} // empty, but not nil
		b, err := marshalJSON(rules)
		require.NoError(t, err)
		assert.Equal(t, []byte("[]"), b)
	})

	t.Run("валидная структура", func(t *testing.T) {
		cfg := &domains.TriggerConfig{ScrollDepth: ptr(50)}
		b, err := marshalJSON(cfg)
		require.NoError(t, err)
		assert.Contains(t, string(b), `"scroll_depth":50`)
	})
}

func ptr[T any](v T) *T {
	return &v
}
