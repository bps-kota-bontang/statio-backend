package repositories

import (
	"statio/internal/models"

	"gorm.io/gorm"
)

type OrganizationRepositoryImpl struct {
	db *gorm.DB
}

// FindByID implements OrganizationRepository.
func (o *OrganizationRepositoryImpl) FindByID(id string) (*models.Organization, error) {
	var org *models.Organization
	if err := o.db.First(&org, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return org, nil
}

// Create implements OrganizationRepository.
func (o *OrganizationRepositoryImpl) Create(org *models.Organization) error {
	if err := o.db.Create(org).Error; err != nil {
		return err
	}
	return nil
}

// Update implements OrganizationRepository.
func (o *OrganizationRepositoryImpl) Update(org *models.Organization) error {
	if err := o.db.Save(org).Error; err != nil {
		return err
	}
	return nil
}

// FindAll implements OrganizationRepository.
func (o *OrganizationRepositoryImpl) FindAll() ([]*models.Organization, error) {
	var organizations []*models.Organization
	if err := o.db.
		Order("name ASC").
		Find(&organizations).Error; err != nil {
		return nil, err
	}
	return organizations, nil
}

func NewOrganizationRepository(db *gorm.DB) OrganizationRepository {
	return &OrganizationRepositoryImpl{db: db}
}
