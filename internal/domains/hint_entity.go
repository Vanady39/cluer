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
	Id            uuid.UUID `json:"id" format:"uuid" example:"01914f6a-8c9b-7a0b-9b0b-8c9b7a0b9b0c"`
	TourVersionId uuid.UUID `json:"tour_version_id" format:"uuid"`
	Step          int       `json:"step" example:"1"`
	Title         string    `json:"title" example:"Нажмите кнопку настроек"`
	Content       string    `json:"content" example:"Здесь вы можете изменить параметры аккаунта"`
	Selector      string    `json:"selector" example:"#settings-btn"`
	Placement     Placement `json:"placement" example:"bottom"`
	// Пусто — подсказка живёт на той же странице, где тур стартовал,
	// то есть наследует target_path версии.
	PagePath         string    `json:"page_path" example:"/dashboard"`
	MediaUrl         string    `json:"media_url" example:"https://cdn.example.com/onboarding/settings.mp4"`
	Spotlight        bool      `json:"spotlight" example:"true"`
	Required         bool      `json:"required" example:"false"`
	WaitForSelector  bool      `json:"wait_for_selector" example:"true"`
	InputPlaceHolder string    `json:"input_placeholder" example:"Введите ваше имя"`
	ExpectedInput    string    `json:"expected_input" example:"John Doe"`
	CreatedAt        time.Time `json:"created_at" format:"date-time"`
	UpdatedAt        time.Time `json:"updated_at" format:"date-time"`
}
