package domains

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlacement_Valid(t *testing.T) {
	tests := []struct {
		name  string
		p     Placement
		valid bool
	}{
		{"top валиден", PlacementTop, true},
		{"bottom валиден", PlacementBottom, true},
		{"center валиден", PlacementCenter, true},
		{"left валиден", PlacementLeft, true},
		{"left-top валиден", PlacementLeftTop, true},
		{"right-top валиден", PlacementRightTop, true},
		{"right валиден", PlacementRight, true},
		{"пустая строка невалидна", Placement(""), false},
		{"произвольное значение невалидно", Placement("diagonal"), false},
		{"top-left невалиден", Placement("top-left"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.p.Valid())
		})
	}
}
