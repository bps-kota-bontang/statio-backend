package repositories

import (
	"statio/internal/models"

	"gorm.io/gorm"
)

type TableRepository interface {
	Count(search string, filters map[string][]string, total *int64) error
	FindAll() ([]*models.Table, error)
	FindPaginated(
		search string,
		limit, offset int,
		sortBy, sortOrder string,
		filters map[string][]string,
	) ([]*models.Table, error)
	CountDimensionsByTableID(tableID string) (*int64, error)
	FindByID(id string) (*models.Table, error)
	FindByIDAndYear(id string, year int) (*models.Table, error)
	FindByIDForFactUpdate(id string) (*models.Table, error)
	CreateWithTx(tx *gorm.DB, table *models.Table) error
}
