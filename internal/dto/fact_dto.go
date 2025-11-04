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
	Dimensions []string `json:"dimensions"` // array of DimensionValue IDs
	Value      *float64 `json:"value"`
}

type UpdateFactRequest struct {
	Year int           `json:"year"`
	Data []FactPayload `json:"data"`
}
