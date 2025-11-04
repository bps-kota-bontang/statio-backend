package repositories

import "statio/internal/models"

type IndicatorRepository interface {
	Count(search string, filters map[string][]string, total *int64) error
	FindPaginated(
		search string,
		limit, offset int,
		sortBy, sortOrder string,
		filters map[string][]string,
	) ([]*models.Indicator, error)
	FindAll() ([]*models.Indicator, error)
	FindAllNames() ([]*models.Indicator, error)
	FindAllMeasures() ([]*models.Indicator, error)
	FindAllUnits() ([]*models.Indicator, error)
	FindByID(id string) (*models.Indicator, error)
	Create(indicator *models.Indicator) error
	Update(indicator *models.Indicator) error
	Delete(id string) error
}
