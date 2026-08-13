package repositories

import (
	"testing"

	"github.com/Vanady39/cluer/onboarding/internal/domains"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalJSON(t *testing.T) {
	t.Run("nil возвращает nil", func(t *testing.T) {
		b, err := marshalJSON(nil)
		assert.NoError(t, err)
		assert.Nil(t, b)
	})

	t.Run("typed nil *triggerConfigRow возвращает nil (не JSONB null)", func(t *testing.T) {
		var row *triggerConfigRow
		b, err := marshalJSON(row)
		require.NoError(t, err)
		assert.Nil(t, b, "typed nil должен давать SQL NULL, а не JSONB null")
	})

	t.Run("валидная структура", func(t *testing.T) {
		row := &triggerConfigRow{ScrollDepth: ptr(50)}
		b, err := marshalJSON(row)
		require.NoError(t, err)
		assert.Contains(t, string(b), `"scroll_depth":50`)
	})
}

func TestUnmarshalJSONB(t *testing.T) {
	t.Run("пустые данные", func(t *testing.T) {
		var row triggerConfigRow
		err := unmarshalJSONB(nil, &row)
		assert.NoError(t, err)
	})

	t.Run("битый JSON оборачивается в ErrCorruptedData", func(t *testing.T) {
		var row triggerConfigRow
		err := unmarshalJSONB([]byte("{"), &row)
		require.Error(t, err)
		assert.ErrorIs(t, err, domains.ErrCorruptedData)
	})
}

// Формат колонок trigger_config и audience_rules закреплён здесь: доменные
// структуры больше не несут json-тегов, поэтому единственное, что защищает уже
// записанные строки от переименования поля в домене, — эти проверки.
func TestTriggerConfigColumnFormat(t *testing.T) {
	t.Run("typed nil даёт SQL NULL", func(t *testing.T) {
		var cfg *domains.TriggerConfig
		b, err := marshalTriggerConfig(cfg)
		require.NoError(t, err)
		assert.Nil(t, b)
	})

	t.Run("имена полей в колонке не изменились", func(t *testing.T) {
		cfg := &domains.TriggerConfig{
			ScrollDepth:     ptr(50),
			InactivitySecs:  ptr(30),
			ElementSelector: ptr("#hero"),
			DelayMs:         ptr(1500),
		}
		b, err := marshalTriggerConfig(cfg)
		require.NoError(t, err)
		assert.JSONEq(t,
			`{"scroll_depth":50,"inactivity_secs":30,"element_selector":"#hero","delay_ms":1500}`,
			string(b),
		)
	})

	t.Run("пустой конфиг не пишет ключей", func(t *testing.T) {
		b, err := marshalTriggerConfig(&domains.TriggerConfig{})
		require.NoError(t, err)
		assert.JSONEq(t, `{}`, string(b))
	})

	t.Run("round-trip через колонку сохраняет значения", func(t *testing.T) {
		cfg := &domains.TriggerConfig{ScrollDepth: ptr(50), DelayMs: ptr(1500)}
		b, err := marshalTriggerConfig(cfg)
		require.NoError(t, err)

		var got *domains.TriggerConfig
		require.NoError(t, unmarshalTriggerConfig(b, &got))
		assert.Equal(t, cfg, got)
	})

	t.Run("NULL оставляет получателя нетронутым", func(t *testing.T) {
		var got *domains.TriggerConfig
		require.NoError(t, unmarshalTriggerConfig(nil, &got))
		assert.Nil(t, got)
	})
}

func TestAudienceRulesColumnFormat(t *testing.T) {
	t.Run("typed nil слайс даёт SQL NULL", func(t *testing.T) {
		var rules []domains.AudienceRule
		b, err := marshalAudienceRules(rules)
		require.NoError(t, err)
		assert.Nil(t, b)
	})

	t.Run("пустой (не nil) слайс даёт []", func(t *testing.T) {
		b, err := marshalAudienceRules([]domains.AudienceRule{})
		require.NoError(t, err)
		assert.Equal(t, []byte("[]"), b)
	})

	t.Run("имена полей в колонке не изменились", func(t *testing.T) {
		rules := []domains.AudienceRule{{
			Type:      "page_visited",
			Key:       "is_premium",
			Operator:  "eq",
			Value:     true,
			Timeframe: "7d",
		}}
		b, err := marshalAudienceRules(rules)
		require.NoError(t, err)
		assert.JSONEq(t,
			`[{"type":"page_visited","key":"is_premium","operator":"eq","value":true,"timeframe":"7d"}]`,
			string(b),
		)
	})

	t.Run("round-trip через колонку сохраняет значения", func(t *testing.T) {
		rules := []domains.AudienceRule{{Type: "user_property", Key: "plan", Operator: "eq", Value: "pro"}}
		b, err := marshalAudienceRules(rules)
		require.NoError(t, err)

		var got []domains.AudienceRule
		require.NoError(t, unmarshalAudienceRules(b, &got))
		assert.Equal(t, rules, got)
	})

	t.Run("NULL оставляет получателя нетронутым", func(t *testing.T) {
		var got []domains.AudienceRule
		require.NoError(t, unmarshalAudienceRules(nil, &got))
		assert.Nil(t, got)
	})
}

func ptr[T any](v T) *T {
	return &v
}
