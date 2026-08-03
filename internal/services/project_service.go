package services

import (
	"statio/internal/dto"
	"statio/internal/mappers"
	"statio/internal/repositories"
	"strings"
)

type ProjectService struct {
	projectRepo repositories.ProjectRepository
}

func NewProjectService(projectRepo repositories.ProjectRepository) *ProjectService {
	return &ProjectService{projectRepo: projectRepo}
}

func (s *ProjectService) GetAllPaginated(search string, page, perPage int, sortBy, sortOrder string) ([]*dto.ProjectListResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	var total int64
	if err := s.projectRepo.Count(search, &total); err != nil {
		return nil, 0, err
	}
	projects, err := s.projectRepo.FindPaginated(search, perPage, (page-1)*perPage, sortBy, sortOrder)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*dto.ProjectListResponse, 0, len(projects))
	for _, project := range projects {
		responses = append(responses, mappers.ToProjectListResponse(project))
	}
	return responses, total, nil
}

func (s *ProjectService) GetByID(id string) (*dto.ProjectResponse, error) {
	project, err := s.projectRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return mappers.ToProjectResponse(project), nil
}

func (s *ProjectService) Create(input *dto.CreateProjectRequest) (*dto.ProjectResponse, error) {
	input.Name = strings.TrimSpace(input.Name)
	project := mappers.ToProjectModel(input)
	if err := s.projectRepo.Create(project); err != nil {
		return nil, err
	}
	return mappers.ToProjectResponse(project), nil
}

func (s *ProjectService) Update(id string, input *dto.UpdateProjectRequest) (*dto.ProjectResponse, error) {
	input.Name = strings.TrimSpace(input.Name)
	project, err := s.projectRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	mappers.ApplyProjectUpdateFromRequest(project, input)
	if err := s.projectRepo.Update(project); err != nil {
		return nil, err
	}
	return mappers.ToProjectResponse(project), nil
}

func (s *ProjectService) Delete(id string) error {
	return s.projectRepo.Delete(id)
}

func (s *ProjectService) AddTable(projectID string, input *dto.ProjectTableRequest) error {
	normalizeProjectTableRequest(input)
	return s.projectRepo.AddTable(projectID, input)
}

func (s *ProjectService) UpdateTable(projectID, tableID string, input *dto.ProjectTableRequest) error {
	normalizeProjectTableRequest(input)
	return s.projectRepo.UpdateTable(projectID, tableID, input)
}

func (s *ProjectService) RemoveTable(projectID, tableID string) error {
	return s.projectRepo.RemoveTable(projectID, tableID)
}

func normalizeProjectTableRequest(input *dto.ProjectTableRequest) {
	input.Page = strings.TrimSpace(input.Page)
	input.TableNumber = strings.TrimSpace(input.TableNumber)
}
