package mappers

import (
	"statio/internal/dto"
	"statio/internal/models"
)

// ToDimensionResponse mengubah models.Dimension menjadi dto.DimensionResponse
func ToDimensionResponse(dimension *models.Dimension) *dto.DimensionResponse {
	return &dto.DimensionResponse{
		ID:   dimension.ID,
		Name: dimension.Name,
		Values: func() []dto.DimensionValueResponse {
			values := make([]dto.DimensionValueResponse, 0, len(dimension.Values))
			for _, v := range dimension.Values {
				values = append(values, *ToDimensionValueResponse(&v))
			}
			return values
		}(),
	}
}

// ToDimensionNameResponse mengubah models.Dimension menjadi dto.DimensionNameResponse
func ToDimensionNameResponse(dimension *models.Dimension) *dto.DimensionNameResponse {
	return &dto.DimensionNameResponse{
		Name: dimension.Name,
	}
}

// ToDimensionValueResponse mengubah models.DimensionValue menjadi dto.DimensionValueResponse
func ToDimensionValueResponse(dv *models.DimensionValue) *dto.DimensionValueResponse {
	return &dto.DimensionValueResponse{
		ID:    dv.ID,
		Name:  dv.Name,
		Order: dv.Order,
	}
}

// ToDimensionModel mengubah dto.CreateDimensionRequest menjadi models.Dimension
func ToDimensionModel(input *dto.CreateDimensionRequest) *models.Dimension {
	return &models.Dimension{
		Name: input.Name,
		Values: func() []models.DimensionValue {
			values := make([]models.DimensionValue, 0, len(input.Values))
			for i, v := range input.Values {
				values = append(values, models.DimensionValue{
					Name:  v,
					Order: i + 1,
				})
			}
			return values
		}(),
	}
}

// ApplyDimensionUpdateFromRequest memperbarui models.Dimension dari dto.UpdateDimensionRequest
func ApplyDimensionUpdateFromRequest(dimension *models.Dimension, input *dto.UpdateDimensionRequest) {
	dimension.Name = input.Name

	// Buat map ID -> Value lama
	existingValues := make(map[string]*models.DimensionValue)
	for i := range dimension.Values {
		existingValues[dimension.Values[i].ID] = &dimension.Values[i]
	}

	var updatedValues []models.DimensionValue

	for idx, v := range input.Values {
		if v.ID != nil && *v.ID != "" {
			// update value lama jika ada
			if existing, ok := existingValues[*v.ID]; ok {
				if v.Name != "" {
					existing.Name = v.Name
				}
				existing.Order = idx
				updatedValues = append(updatedValues, *existing)
			} else {
				// ID tidak ditemukan, bisa dianggap error atau ignore
				// fmt.Println("Warning: ID value tidak ditemukan:", *v.ID)
			}
		} else if v.Name != "" {
			// value baru
			updatedValues = append(updatedValues, models.DimensionValue{
				Name:  v.Name,
				Order: idx,
			})
		}
	}

	dimension.Values = updatedValues
}
