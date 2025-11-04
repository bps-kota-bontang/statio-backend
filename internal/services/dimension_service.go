package services

import (
	"fmt"
	"statio/internal/dto"
	"statio/internal/mappers"
	"statio/internal/repositories"
)

type DimensionService struct {
	dimensionRepo repositories.DimensionRepository
}

func NewDimensionService(dimensionRepo repositories.DimensionRepository) *DimensionService {
	return &DimensionService{dimensionRepo: dimensionRepo}
}

func (s *DimensionService) GetAll() ([]*dto.DimensionResponse, error) {
	dimensions, err := s.dimensionRepo.FindAll()
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.DimensionResponse, 0, len(dimensions))
	for _, dimension := range dimensions {
		responses = append(responses, mappers.ToDimensionResponse(dimension))
	}

	return responses, nil
}

func (s *DimensionService) GetAllPaginated(
	search string,
	page, perPage int,
	sortBy, sortOrder string,
	filters map[string][]string,
) ([]*dto.DimensionResponse, int64, error) {

	var total int64
	if err := s.dimensionRepo.Count(search, filters, &total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	dimensions, err := s.dimensionRepo.FindPaginated(search, perPage, offset, sortBy, sortOrder, filters)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*dto.DimensionResponse, 0, len(dimensions))
	for _, dimension := range dimensions {
		responses = append(responses, mappers.ToDimensionResponse(dimension))
	}

	return responses, total, nil
}

func (s *DimensionService) GetAllNames() ([]*dto.DimensionNameResponse, error) {
	dimensions, err := s.dimensionRepo.FindAllNames()
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.DimensionNameResponse, 0, len(dimensions))
	for _, dimension := range dimensions {
		responses = append(responses, mappers.ToDimensionNameResponse(dimension))
	}

	return responses, nil
}

func (s *DimensionService) GetByID(id string) (*dto.DimensionResponse, error) {
	dimension, err := s.dimensionRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	response := mappers.ToDimensionResponse(dimension)
	return response, nil
}

func (s *DimensionService) Create(input *dto.CreateDimensionRequest) (*dto.DimensionResponse, error) {
	dimension := mappers.ToDimensionModel(input)
	if err := s.dimensionRepo.Create(dimension); err != nil {
		return nil, err
	}

	response := mappers.ToDimensionResponse(dimension)
	return response, nil
}

func (s *DimensionService) Update(id string, input *dto.UpdateDimensionRequest) (*dto.DimensionResponse, error) {
	dimension, err := s.dimensionRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	mappers.ApplyDimensionUpdateFromRequest(dimension, input)
	if err := s.dimensionRepo.Update(dimension); err != nil {
		return nil, err
	}

	response := mappers.ToDimensionResponse(dimension)
	return response, nil
}

func (s *DimensionService) ValidateIDs(ids []string) error {
	valid, err := s.dimensionRepo.AreValidIDs(ids)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("some dimension IDs are invalid")
	}
	return nil
}
