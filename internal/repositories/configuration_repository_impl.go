package repositories

import (
	"statio/internal/models"

	"gorm.io/gorm"
)

type ConfigurationRepositoryImpl struct {
	db *gorm.DB
}

func NewConfigurationRepository(db *gorm.DB) ConfigurationRepository {
	return &ConfigurationRepositoryImpl{db: db}
}

// Create implements [ConfigurationRepository].
func (c *ConfigurationRepositoryImpl) Create(configuration *models.Configuration) error {
	if err := c.db.Create(configuration).Error; err != nil {
		return err
	}
	return nil
}

// FindByKey implements [ConfigurationRepository].
func (c *ConfigurationRepositoryImpl) FindByKey(key string) (*models.Configuration, error) {
	var configuration models.Configuration
	if err := c.db.Where("key = ?", key).First(&configuration).Error; err != nil {
		return nil, err
	}
	return &configuration, nil
}

// Update implements [ConfigurationRepository].
func (c *ConfigurationRepositoryImpl) Update(configuration *models.Configuration) error {
	if err := c.db.Save(configuration).Error; err != nil {
		return err
	}
	return nil
}
