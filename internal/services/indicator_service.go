package services

import (
	"statio/internal/dto"
	"statio/internal/mappers"
	"statio/internal/repositories"
)

type IndicatorService struct {
	indicatorRepo repositories.IndicatorRepository
}

func NewIndicatorService(indicatorRepo repositories.IndicatorRepository) *IndicatorService {
	return &IndicatorService{indicatorRepo: indicatorRepo}
}

func (s *IndicatorService) GetAll() ([]*dto.IndicatorResponse, error) {
	indicators, err := s.indicatorRepo.FindAll()
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.IndicatorResponse, 0, len(indicators))
	for _, indicator := range indicators {
		responses = append(responses, mappers.ToIndicatorResponse(indicator))
	}

	return responses, nil
}

func (s *IndicatorService) GetAllNames() ([]*dto.IndicatorNameResponse, error) {
	names, err := s.indicatorRepo.FindAllNames()
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.IndicatorNameResponse, 0, len(names))
	for _, name := range names {
		responses = append(responses, mappers.ToIndicatorNameResponse(name))
	}

	return responses, nil
}

func (s *IndicatorService) GetAllMeasures() ([]*dto.IndicatorMeasureResponse, error) {
	measures, err := s.indicatorRepo.FindAllMeasures()
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.IndicatorMeasureResponse, 0, len(measures))
	for _, measure := range measures {
		responses = append(responses, mappers.ToIndicatorMeasureResponse(measure))
	}

	return responses, nil
}

func (s *IndicatorService) GetAllUnits() ([]*dto.IndicatorUnitResponse, error) {
	units, err := s.indicatorRepo.FindAllUnits()
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.IndicatorUnitResponse, 0, len(units))
	for _, unit := range units {
		responses = append(responses, mappers.ToIndicatorUnitResponse(unit))
	}

	return responses, nil
}

func (s *IndicatorService) GetAllPaginated(
	search string,
	page, perPage int,
	sortBy, sortOrder string,
	filters map[string][]string,
) ([]*dto.IndicatorResponse, int64, error) {

	var total int64
	if err := s.indicatorRepo.Count(search, filters, &total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	indicators, err := s.indicatorRepo.FindPaginated(search, perPage, offset, sortBy, sortOrder, filters)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*dto.IndicatorResponse, 0, len(indicators))
	for _, indicator := range indicators {
		responses = append(responses, mappers.ToIndicatorResponse(indicator))
	}

	return responses, total, nil
}

func (s *IndicatorService) GetByID(id string) (*dto.IndicatorResponse, error) {
	indicator, err := s.indicatorRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	response := mappers.ToIndicatorResponse(indicator)
	return response, nil
}

func (s *IndicatorService) Create(input *dto.CreateIndicatorRequest) (*dto.IndicatorResponse, error) {
	indicator := mappers.ToIndicatorModel(input)
	if err := s.indicatorRepo.Create(indicator); err != nil {
		return nil, err
	}

	response := mappers.ToIndicatorResponse(indicator)
	return response, nil
}

func (s *IndicatorService) Update(id string, input *dto.UpdateIndicatorRequest) (*dto.IndicatorResponse, error) {
	indicator, err := s.indicatorRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	mappers.ApplyIndicatorUpdateFromRequest(indicator, input)
	if err := s.indicatorRepo.Update(indicator); err != nil {
		return nil, err
	}

	response := mappers.ToIndicatorResponse(indicator)
	return response, nil
}
