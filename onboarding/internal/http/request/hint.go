package request

import (
	"time"

	"github.com/Vanady39/cluer/onboarding/internal/domains"
	"github.com/google/uuid"
)

type Hint struct {
	Id            uuid.UUID `json:"id" format:"uuid"`
	TourVersionId uuid.UUID `json:"tour_version_id" format:"uuid"`
	Step          int       `json:"step" example:"1"`
	Title         string    `json:"title" example:"Нажмите кнопку настроек"`
	Content       string    `json:"content" example:"Здесь вы можете изменить параметры аккаунта"`
	Selector      string    `json:"selector" example:"#settings-btn"`
	Placement     string    `json:"placement" example:"bottom" enums:"top,bottom,center,left-top,right-top,left,right"`
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

type ReorderRequest struct {
	HintIds []uuid.UUID `json:"hint_ids" binding:"required"`
}

func (h *Hint) ToDomain() *domains.Hint {
	if h == nil {
		return nil
	}
	out := h.toDomainValue()
	return &out
}

func (h *Hint) toDomainValue() domains.Hint {
	return domains.Hint{
		Id:               h.Id,
		TourVersionId:    h.TourVersionId,
		Step:             h.Step,
		Title:            h.Title,
		Content:          h.Content,
		Selector:         h.Selector,
		Placement:        domains.Placement(h.Placement),
		PagePath:         h.PagePath,
		MediaUrl:         h.MediaUrl,
		Spotlight:        h.Spotlight,
		Required:         h.Required,
		WaitForSelector:  h.WaitForSelector,
		InputPlaceHolder: h.InputPlaceHolder,
		ExpectedInput:    h.ExpectedInput,
		CreatedAt:        h.CreatedAt,
		UpdatedAt:        h.UpdatedAt,
	}
}

func HintsToDomain(hints []Hint) []domains.Hint {
	if hints == nil {
		return nil
	}
	out := make([]domains.Hint, 0, len(hints))
	for i := range hints {
		out = append(out, hints[i].toDomainValue())
	}
	return out
}
