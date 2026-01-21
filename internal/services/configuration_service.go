package services

import (
	"statio/internal/dto"
	"statio/internal/mappers"
	"statio/internal/models"
	"statio/internal/repositories"
)

type ConfigurationService struct {
	repository repositories.ConfigurationRepository
}

func NewConfigurationService(repository repositories.ConfigurationRepository) *ConfigurationService {
	return &ConfigurationService{repository: repository}
}

func (c *ConfigurationService) GetConfigurationByKey(key string) (*dto.ConfigurationResponse, error) {
	configuration, err := c.repository.FindByKey(key)
	if err != nil {
		return nil, err
	}

	resp := mappers.ToConfigurationResponse(configuration)
	return resp, nil
}

func (c *ConfigurationService) UpdateConfiguration(key string, req *dto.UpdateConfigurationRequest) error {

	configurationExists, _ := c.repository.FindByKey(key)

	var configuration *models.Configuration

	if configurationExists == nil {
		configuration = &models.Configuration{
			Key:   key,
			Name:  req.Name,
			Value: req.Value,
		}
	} else {
		configurationExists.Name = req.Name
		configurationExists.Value = req.Value
		configuration = configurationExists
	}

	if err := c.repository.Update(configuration); err != nil {
		return err
	}

	return nil
}
