package domains

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTriggerType_Valid(t *testing.T) {
	tests := []struct {
		name  string
		tt    TriggerType
		valid bool
	}{
		{"on_load валиден", TriggerOnLoad, true},
		{"delay валиден", TriggerDelay, true},
		{"exit_intent валиден", TriggerExitIntent, true},
		{"manual валиден", TriggerManual, true},
		{"scroll_depth валиден", TriggerScrollDepth, true},
		{"inactivity валиден", TriggerInactivity, true},
		{"element_visible валиден", TriggerElementVisible, true},
		{"пустая строка невалидна", TriggerType(""), false},
		{"произвольное значение невалидно", TriggerType("on_click"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.tt.Valid())
		})
	}
}
