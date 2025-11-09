package repositories

import "statio/internal/models"

type OrganizationRepository interface {
	FindAll() ([]*models.Organization, error)
	FindByID(id string) (*models.Organization, error)
	Create(org *models.Organization) error
	Update(org *models.Organization) error
}
