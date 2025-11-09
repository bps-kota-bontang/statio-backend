package mappers

import (
	"statio/internal/dto"
	"statio/internal/models"
)

func ToOrganizationResponse(org *models.Organization) *dto.OrganizationResponse {
	return &dto.OrganizationResponse{
		ID:   org.ID,
		Name: org.Name,
	}
}
