package repositories

import (
	"statio/internal/dto"
	"statio/internal/models"

	"gorm.io/gorm"
)

type TableRepository interface {
	Count(search string, filters map[string][]string, total *int64) error
	FindLight(search string, sortBy string, sortOrder string, filters map[string][]string) ([]*models.Table, error)
	LoadDimensionsForTableIDs(ids []string) ([]*models.Table, error)
	FindByIDsDetailed(ids []string) ([]*models.Table, error)
	FindAll() ([]*models.Table, error)
	FindPaginated(
		search string,
		limit, offset int,
		sortBy, sortOrder string,
		filters map[string][]string,
	) ([]*models.Table, error)
	CountDimensionsByTableID(tableID string) (*int64, error)
	FindBaseByID(id string) (*models.Table, error)
	FindDetailedByID(id string) (*models.Table, error)
	FindByIDAndYear(id string, year int) (*models.Table, error)
	FindForFactUpdate(id string) (*models.Table, error)
	CreateWithTx(tx *gorm.DB, table *models.Table) error
	UpdateWithRelations(table *models.Table, dimensionIDs []string) error
	Update(table *models.Table) error
	UpdateOrganizationBulk(organizationID string, tableIDs []string) error
	AddLabelsBulk(labels []string, tableIDs []string) error
	UpdateLabels(tableID string, labels []string) error
	FindAllLabels() ([]*string, error)
	FindByIDs(tableIDs []string) ([]*models.Table, error)
	FindTablesBase(filter *dto.FilterTablesRequest) ([]*models.Table, error)
}
