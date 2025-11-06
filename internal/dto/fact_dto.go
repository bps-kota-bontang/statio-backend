package dto

type FactDimensionResponse struct {
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Value DimensionValueResponse `json:"value"`
}

type FactResponse struct {
	Value      *float64                `json:"value"`
	Year       int                     `json:"year"`
	Dimensions []FactDimensionResponse `json:"dimensions"`
}

type FactPayload struct {
	Dimensions []string `json:"dimensions" validate:"required,min=0,max=2,dive,required"`
	Value      *float64 `json:"value"`
	Year       int      `json:"year"`
}

type UpdateFactRequest struct {
	Data []FactPayload `json:"data"`
}
