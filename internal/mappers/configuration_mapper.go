package mappers

import (
	"statio/internal/dto"
	"statio/internal/models"
)

func ToConfigurationResponse(configuration *models.Configuration) *dto.ConfigurationResponse {
	return &dto.ConfigurationResponse{
		ID:    configuration.ID,
		Name:  configuration.Name,
		Key:   configuration.Key,
		Value: configuration.Value,
	}
}
