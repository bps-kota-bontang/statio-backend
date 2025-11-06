package dto

type TableResponse struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Direction   int                 `json:"direction"`
	Description *string             `json:"description,omitempty"`
	Indicator   *IndicatorResponse  `json:"indicator"`
	Dimensions  []DimensionResponse `json:"dimensions"`
	Facts       []FactResponse      `json:"facts"`
}

type TableListResponse struct {
	ID         string                `json:"id"`
	Name       string                `json:"name"`
	Indicator  IndicatorListResponse `json:"indicator"`
	Dimensions []string              `json:"dimensions"`
}

type CreateTableRequest struct {
	Name         string   `json:"name" validate:"required"`
	IndicatorID  string   `json:"indicator_id" validate:"required"`
	DimensionIDs []string `json:"dimension_ids" validate:"required,min=0,max=2,dive,required"`
}
