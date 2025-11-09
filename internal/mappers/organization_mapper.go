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

func ToOrganizationModel(input *dto.CreateOrganizationRequest) *models.Organization {
	return &models.Organization{
		Name: input.Name,
	}
}

func ApplyOrganizationUpdateFromRequest(org *models.Organization, input *dto.UpdateOrganizationRequest) {
	if input.Name != nil {
		org.Name = *input.Name
	}
}
