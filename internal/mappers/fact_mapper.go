package mappers

import (
	"statio/internal/dto"
	"statio/internal/models"
)

func TransformFactDimension(dv *models.DimensionValue) *dto.FactDimensionResponse {
	return &dto.FactDimensionResponse{
		ID:    dv.Dimension.ID,
		Name:  dv.Dimension.Name,
		Value: *ToDimensionValueResponse(dv),
	}
}

func ToFactResponse(fact *models.Fact) *dto.FactResponse {
	dimensions := make([]dto.FactDimensionResponse, 0, len(fact.FactDimensionValues))
	for _, fdv := range fact.FactDimensionValues {
		if fdv.DimensionValue != nil {
			dimensions = append(dimensions, *TransformFactDimension(fdv.DimensionValue))
		}
	}

	return &dto.FactResponse{
		OldValue:   fact.OldValue,
		Value:      fact.Value,
		Year:       fact.Year,
		IsOutlier:  fact.IsOutlier,
		Dimensions: dimensions,
	}
}
