package mappers

import (
	"statio/internal/dto"
	"statio/internal/models"
)

// ToIndicatorResponse mengubah models.Indicator menjadi dto.IndicatorResponse
func ToIndicatorResponse(indicator *models.Indicator) *dto.IndicatorResponse {
	return &dto.IndicatorResponse{
		ID:      indicator.ID,
		Name:    indicator.Name,
		Measure: indicator.Measure,
		Unit:    indicator.Unit,
	}
}

// ToIndicatorNameResponse mengubah models.Indicator menjadi dto.IndicatorNameResponse
func ToIndicatorNameResponse(indicator *models.Indicator) *dto.IndicatorNameResponse {
	return &dto.IndicatorNameResponse{
		Name: indicator.Name,
	}
}

// ToIndicatorMeasureResponse mengubah models.Indicator menjadi dto.IndicatorMeasureResponse
func ToIndicatorMeasureResponse(indicator *models.Indicator) *dto.IndicatorMeasureResponse {
	return &dto.IndicatorMeasureResponse{
		Measure: indicator.Measure,
	}
}

// ToIndicatorUnitResponse mengubah models.Indicator menjadi dto.IndicatorUnitResponse
func ToIndicatorUnitResponse(indicator *models.Indicator) *dto.IndicatorUnitResponse {
	return &dto.IndicatorUnitResponse{
		Unit: *indicator.Unit,
	}
}

// TransformCreateIndicatorRequest mengubah dto.CreateIndicatorRequest menjadi models.Indicator
func ToIndicatorModel(input *dto.CreateIndicatorRequest) *models.Indicator {
	return &models.Indicator{
		Name:    input.Name,
		Measure: input.Measure,
		Unit:    input.Unit,
	}
}

// ApplyIndicatorUpdateFromRequest memperbarui models.Indicator dari dto.UpdateIndicatorRequest
func ApplyIndicatorUpdateFromRequest(indicator *models.Indicator, input *dto.UpdateIndicatorRequest) {
	if input.Name != nil {
		indicator.Name = *input.Name
	}
	if input.Measure != nil {
		indicator.Measure = *input.Measure
	}
	if input.Unit != nil {
		indicator.Unit = input.Unit
	}
}

// ToIndicatorListResponse mengubah models.Indicator menjadi dto.IndicatorListResponse
func ToIndicatorListResponse(indicator *models.Indicator) *dto.IndicatorListResponse {
	return &dto.IndicatorListResponse{
		Name:    indicator.Name,
		Measure: indicator.Measure,
		Unit:    indicator.Unit,
	}
}
