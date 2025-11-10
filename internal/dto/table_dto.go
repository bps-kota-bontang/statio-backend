package dto

type TableResponse struct {
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Direction    int                   `json:"direction"`
	Description  *string               `json:"description,omitempty"`
	Indicator    *IndicatorResponse    `json:"indicator"`
	Organization *OrganizationResponse `json:"organization"`
	Labels       []string              `json:"labels"`
	Dimensions   []DimensionResponse   `json:"dimensions"`
	Facts        []FactResponse        `json:"facts"`
}

type TableListResponse struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Indicator    *IndicatorListResponse `json:"indicator"`
	Organization *OrganizationResponse  `json:"organization"`
	Labels       []string               `json:"labels"`
	Dimensions   []string               `json:"dimensions"`
}

type CreateTableRequest struct {
	Name           string   `json:"name" validate:"required"`
	IndicatorID    string   `json:"indicator_id" validate:"required"`
	OrganizationID *string  `json:"organization_id"`
	DimensionIDs   []string `json:"dimension_ids" validate:"required,min=0,max=2,dive,required"`
}

type UpdateTableRequest struct {
	Name           *string  `json:"name,omitempty"`
	IndicatorID    *string  `json:"indicator_id,omitempty"`
	OrganizationID *string  `json:"organization_id,omitempty"`
	DimensionIDs   []string `json:"dimension_ids,omitempty"`
}

type AddLabelsToTablesRequest struct {
	Labels   []string `json:"labels" validate:"required,min=1,dive,required"`
	TableIDs []string `json:"table_ids" validate:"required,min=1,dive,required"`
}

type TableLabelResponse struct {
	Name string `json:"name"`
}

type UpdateTableLabelsRequest struct {
	Labels []string `json:"labels" validate:"required,min=0,dive,required"`
}
