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
	Id               uuid.UUID `json:"id" db:"id" example:"01914f6a-8c9b-7a0b-9b0b-8c9b7a0b9b0c"`
	TourId           uuid.UUID `json:"tour_id" db:"tour_id" example:"01914f6a-8c9b-7a0b-9b0b-8c9b7a0b9b0b"`
	Step             int       `json:"step" db:"step" example:"1"`
	Title            string    `json:"title" db:"title" example:"Нажмите кнопку настроек"`
	Content          string    `json:"content" db:"content" example:"Здесь вы можете изменить параметры аккаунта"`
	Selector         string    `json:"selector" db:"selector" example:"#settings-btn"`
	Placement        Placement `json:"placement" db:"placement" example:"bottom"`
	MediaUrl         string    `json:"media_url" db:"media_url" example:"https://cdn.example.com/onboarding/settings.mp4"`
	Spotlight        bool      `json:"spotlight" db:"spotlight" example:"true"`
	Required         bool      `json:"required" db:"required" example:"false"`
	WaitForSelector  bool      `json:"wait_for_selector" db:"wait_for_selector" example:"true"`
	InputPlaceHolder string    `json:"input_placeholder" db:"input_placeholder" example:"Введите ваше имя"`
	ExpectedInput    string    `json:"expected_input" db:"expected_input" example:"John Doe"`
	CreatedAt        time.Time `json:"created_at" db:"created_at" format:"date-time" example:"2026-08-05T12:00:00Z"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at" format:"date-time" example:"2026-08-05T12:00:00Z"`
}
