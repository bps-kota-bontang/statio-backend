package dto

type IndicatorResponse struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Measure string  `json:"measure"`
	Unit    *string `json:"unit,omitempty"`
}

type IndicatorListResponse struct {
	Name    string  `json:"name"`
	Measure string  `json:"measure"`
	Unit    *string `json:"unit"`
}

type IndicatorNameResponse struct {
	Name string `json:"name"`
}

type IndicatorMeasureResponse struct {
	Measure string `json:"measure"`
}

type IndicatorUnitResponse struct {
	Unit string `json:"unit"`
}

type CreateIndicatorRequest struct {
	Name    string  `json:"name" validate:"required"`
	Measure string  `json:"measure" validate:"required"`
	Unit    *string `json:"unit,omitempty"`
}

type UpdateIndicatorRequest struct {
	Name    *string `json:"name,omitempty"`
	Measure *string `json:"measure,omitempty"`
	Unit    *string `json:"unit,omitempty"`
}
