package mappers

import (
	"statio/internal/dto"
	"statio/internal/models"
)

func ToProjectListResponse(project *models.Project) *dto.ProjectListResponse {
	return &dto.ProjectListResponse{
		ID:         project.ID,
		Name:       project.Name,
		TableCount: len(project.Tables),
	}
}

func ToProjectResponse(project *models.Project) *dto.ProjectResponse {
	response := &dto.ProjectResponse{
		ID:     project.ID,
		Name:   project.Name,
		Tables: make([]dto.ProjectTableResponse, 0, len(project.Tables)),
	}
	for _, relation := range project.Tables {
		if relation.Table == nil {
			continue
		}
		response.Tables = append(response.Tables, dto.ProjectTableResponse{
			TableID:     relation.TableID,
			TableName:   relation.Table.Name,
			Page:        relation.Page,
			TableNumber: relation.TableNumber,
		})
	}
	return response
}

func ToProjectModel(input *dto.CreateProjectRequest) *models.Project {
	return &models.Project{Name: input.Name}
}

func ApplyProjectUpdateFromRequest(project *models.Project, input *dto.UpdateProjectRequest) {
	project.Name = input.Name
}
