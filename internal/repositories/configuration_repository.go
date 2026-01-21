package repositories

import "statio/internal/models"

type ConfigurationRepository interface {
	Update(configuration *models.Configuration) error
	Create(configuration *models.Configuration) error
	FindByKey(key string) (*models.Configuration, error)
}
