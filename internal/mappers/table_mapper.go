package mappers

import (
	"statio/internal/dto"
	"statio/internal/models"
	"statio/utils"
)

// ToTableResponse mengubah models.Table menjadi dto.TableResponse
func ToTableResponse(table *models.Table, year *int) *dto.TableResponse {
	resp := &dto.TableResponse{
		ID:          table.ID,
		Name:        table.Name,
		Direction:   table.Direction,
		Description: table.Description,
		Dimensions:  []dto.DimensionResponse{},
		Facts:       []dto.FactResponse{},
	}

	// Transform Dimensions
	dimensionNames := []string{}
	for _, td := range table.Dimensions {
		if td.Dimension == nil {
			continue
		}

		resp.Dimensions = append(resp.Dimensions, *ToDimensionResponse(td.Dimension))

		dimensionNames = append(dimensionNames, td.Dimension.Name)
	}

	// Transform Indicator (hanya 1 indikator)
	if table.Indicator != nil {
		resp.Indicator = ToIndicatorResponse(table.Indicator)
	}

	// Transform Facts untuk indikator ini dan tahun tertentu
	for _, f := range table.Facts {
		if year != nil {
			if f.Year != *year {
				continue
			}
		}
		fr := ToFactResponse(&f)

		// Pastikan semua dimension tercakup
		complete := true
		for _, dimName := range dimensionNames {
			found := false
			for _, v := range fr.Dimensions {
				if v.Name == dimName {
					found = true
					break
				}
			}
			if !found {
				complete = false
				break
			}
		}
		if !complete {
			continue
		}

		resp.Facts = append(resp.Facts, *fr)
	}

	if year != nil {
		resp.Facts = TransformFactsWithBlanks(table, *year)
	}

	return resp
}

func TransformFactsWithBlanks(table *models.Table, year int) []dto.FactResponse {
	// map existing facts berdasarkan key aman
	existingFacts := map[string]*models.Fact{}
	for i := range table.Facts {
		f := &table.Facts[i]
		if f.Year != year {
			continue
		}
		if len(f.FactDimensionValues) == 0 {
			continue
		}
		dimIDs := make([]string, len(f.FactDimensionValues))
		for j, fdv := range f.FactDimensionValues {
			dimIDs[j] = fdv.DimensionValueID
		}
		key := utils.DimensionValueKeyFromIDs(dimIDs)
		existingFacts[key] = f
	}

	// list semua nilai dari setiap dimensi
	dimValuesList := [][]models.DimensionValue{}
	for _, td := range table.Dimensions {
		dimValuesList = append(dimValuesList, td.Dimension.Values)
	}

	var facts []dto.FactResponse

	var generate func(idx int, current []dto.FactDimensionResponse)
	generate = func(idx int, current []dto.FactDimensionResponse) {
		if idx == len(dimValuesList) {

			dimIDs := make([]string, len(current))
			for i, d := range current {
				dimIDs[i] = d.Value.ID
			}
			key := utils.DimensionValueKeyFromIDs(dimIDs)

			if f, ok := existingFacts[key]; ok {
				// pakai fact yang sudah ada
				dims := make([]dto.FactDimensionResponse, len(current))
				copy(dims, current)
				facts = append(facts, dto.FactResponse{
					Value:      f.Value,
					Year:       f.Year,
					Dimensions: dims,
				})
			} else {
				// blank fact
				dims := make([]dto.FactDimensionResponse, len(current))
				copy(dims, current)
				facts = append(facts, dto.FactResponse{
					Value:      nil,
					Year:       year,
					Dimensions: dims,
				})
			}
			return
		}

		for _, v := range dimValuesList[idx] {
			dim := table.Dimensions[idx].Dimension
			next := append([]dto.FactDimensionResponse{}, current...)
			next = append(next, dto.FactDimensionResponse{
				ID:    dim.ID,
				Name:  dim.Name,
				Value: *ToDimensionValueResponse(&v),
			})
			generate(idx+1, next)
		}
	}

	generate(0, []dto.FactDimensionResponse{})
	return facts
}

// ToTableListResponse mengubah models.Table menjadi dto.TableListResponse
func ToTableListResponse(table *models.Table) *dto.TableListResponse {
	return &dto.TableListResponse{
		ID:         table.ID,
		Name:       table.Name,
		Indicator:  *ToIndicatorListResponse(table.Indicator),
		Dimensions: extractDimensionNames(table.Dimensions),
	}
}

func extractDimensionNames(dims []models.TableDimension) []string {
	names := make([]string, 0, len(dims))
	for _, td := range dims {
		if td.Dimension != nil {
			names = append(names, td.Dimension.Name)
		}
	}
	return names
}

func ToTableModel(input *dto.CreateTableRequest) *models.Table {
	// Pre-allocate dimensions slice, bisa kosong
	dimensions := make([]models.TableDimension, len(input.DimensionIDs))
	for i, dimID := range input.DimensionIDs {
		dimensions[i] = models.TableDimension{
			DimensionID: dimID,
		}
	}

	table := &models.Table{
		Name:        input.Name,
		Direction:   len(input.DimensionIDs),
		IndicatorID: input.IndicatorID,
		Dimensions:  dimensions,
	}

	return table
}
