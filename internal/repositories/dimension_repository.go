package repositories

import "statio/internal/models"

type DimensionRepository interface {
	Count(search string, filters map[string][]string, total *int64) error
	FindPaginated(
		search string,
		limit, offset int,
		sortBy, sortOrder string,
		filters map[string][]string,
	) ([]*models.Dimension, error)
	FindAll() ([]*models.Dimension, error)
	FindAllNames() ([]*models.Dimension, error)
	FindByID(id string) (*models.Dimension, error)
	FindByIDWithValues(id string) (*models.Dimension, error)
	FindParentValueIDs(dimensionID string) ([]string, error)
	Create(dimension *models.Dimension) error
	Update(dimension *models.Dimension) error
	Delete(id string) error
	AreValidIDs(ids []string) (bool, error)
}
