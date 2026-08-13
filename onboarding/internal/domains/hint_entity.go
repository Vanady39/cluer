package domains

import (
	"time"

	"github.com/google/uuid"
)

type Placement string

const (
	PlacementTop      Placement = "top"
	PlacementBottom   Placement = "bottom"
	PlacementCenter   Placement = "center"
	PlacementLeftTop  Placement = "left-top"
	PlacementRightTop Placement = "right-top"
	PlacementLeft     Placement = "left"
	PlacementRight    Placement = "right"
)

func (p Placement) Valid() bool {
	switch p {
	case PlacementTop, PlacementBottom, PlacementLeft, PlacementLeftTop, PlacementRightTop, PlacementRight, PlacementCenter:
		return true
	}
	return false
}

type Hint struct {
	Id            uuid.UUID
	TourVersionId uuid.UUID
	Step          int
	Title         string
	Content       string
	Selector      string
	Placement     Placement
	// Пусто — подсказка живёт на той же странице, где тур стартовал,
	// то есть наследует target_path версии.
	PagePath         string
	MediaUrl         string
	Spotlight        bool
	Required         bool
	WaitForSelector  bool
	InputPlaceHolder string
	ExpectedInput    string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
