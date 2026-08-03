package repositories

import (
	"errors"
	"statio/internal/dto"
	"statio/internal/models"
)

var ErrProjectTableExists = errors.New("table already belongs to project")

type ProjectRepository interface {
	Count(search string, total *int64) error
	FindPaginated(search string, limit, offset int, sortBy, sortOrder string) ([]*models.Project, error)
	FindByID(id string) (*models.Project, error)
	Create(project *models.Project) error
	Update(project *models.Project) error
	Delete(id string) error
	AddTable(projectID string, input *dto.ProjectTableRequest) error
	UpdateTable(projectID, tableID string, input *dto.ProjectTableRequest) error
	RemoveTable(projectID, tableID string) error
}
