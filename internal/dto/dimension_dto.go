package dto

type ParentDimensionValueResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DimensionValueResponse struct {
	ID     string                        `json:"id"`
	Name   string                        `json:"name"`
	Order  int                           `json:"order"`
	Parent *ParentDimensionValueResponse `json:"parent,omitempty"`
}

type DimensionResponse struct {
	ID     string                   `json:"id"`
	Name   string                   `json:"name"`
	Order  *int                     `json:"order,omitempty"`
	Values []DimensionValueResponse `json:"values"`
}

type DimensionListResponse struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	HasParentDimension bool   `json:"has_parent_dimension"`
}

type DimensionNameResponse struct {
	Name string `json:"name"`
}

type CreateDimensionRequest struct {
	Name   string   `json:"name" validate:"required"`
	Values []string `json:"values" validate:"required,dive,required"`
}

type UpdateDimensionRequest struct {
	Name   string `json:"name" validate:"required"`
	Values []struct {
		ID   *string `json:"id,omitempty"`
		Name string  `json:"name" validate:"required"`
	} `json:"values"`
}
