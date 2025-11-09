package services

import (
	"statio/internal/dto"
	"statio/internal/mappers"
	"statio/internal/repositories"
)

type OrganizationService struct {
	organizationRepo repositories.OrganizationRepository
	tableSvc         *TableService
}

func NewOrganizationService(organizationRepo repositories.OrganizationRepository, tableSvc *TableService) *OrganizationService {
	return &OrganizationService{
		organizationRepo: organizationRepo,
		tableSvc:         tableSvc,
	}
}

// GetAll retrieves all organizations.
func (s *OrganizationService) GetAll() ([]*dto.OrganizationResponse, error) {
	organizations, err := s.organizationRepo.FindAll()
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.OrganizationResponse, 0, len(organizations))
	for _, org := range organizations {
		responses = append(responses, mappers.ToOrganizationResponse(org))
	}

	return responses, nil
}

// Create creates a new organization.
func (s *OrganizationService) Create(req *dto.CreateOrganizationRequest) (*dto.OrganizationResponse, error) {
	org := mappers.ToOrganizationModel(req)
	if err := s.organizationRepo.Create(org); err != nil {
		return nil, err
	}
	return mappers.ToOrganizationResponse(org), nil
}

// Update updates an existing organization.
func (s *OrganizationService) Update(id string, req *dto.UpdateOrganizationRequest) (*dto.OrganizationResponse, error) {
	org, err := s.organizationRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	mappers.ApplyOrganizationUpdateFromRequest(org, req)

	if err := s.organizationRepo.Update(org); err != nil {
		return nil, err
	}

	return mappers.ToOrganizationResponse(org), nil
}

// AssignTablesToOrganization associates tables with an organization.
func (s *OrganizationService) AssignTablesToOrganization(organizationID string, req *dto.AssignTablesRequest) error {
	return s.tableSvc.AssignOrganizationBulk(organizationID, req.TableIDs)
}
