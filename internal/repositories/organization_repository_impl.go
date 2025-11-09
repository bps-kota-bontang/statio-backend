package repositories

import (
	"statio/internal/models"

	"gorm.io/gorm"
)

type OrganizationRepositoryImpl struct {
	db *gorm.DB
}

// FindAll implements OrganizationRepository.
func (o *OrganizationRepositoryImpl) FindAll() ([]*models.Organization, error) {
	var organizations []*models.Organization
	if err := o.db.Find(&organizations).Error; err != nil {
		return nil, err
	}
	return organizations, nil
}

func NewOrganizationRepository(db *gorm.DB) OrganizationRepository {
	return &OrganizationRepositoryImpl{db: db}
}
