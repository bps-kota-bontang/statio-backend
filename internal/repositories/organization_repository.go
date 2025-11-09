package repositories

import "statio/internal/models"

type OrganizationRepository interface {
	FindAll() ([]*models.Organization, error)
}
