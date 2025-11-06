package repositories

import (
	"statio/internal/models"

	"gorm.io/gorm"
)

type FactRepository interface {
	FindFactByDimensionValues(tableID string, year int, dimValueIDs []string) (*models.Fact, error)
	FindAllByTableAndYear(tableID string, year int) ([]*models.Fact, error)
	FindAllByTable(tableID string) ([]*models.Fact, error)
	CreateFactsTx(tx *gorm.DB, facts []models.Fact) error
	UpdateFactsTx(tx *gorm.DB, facts []*models.Fact) error
	CreateFactDimensionValuesTx(tx *gorm.DB, fdvs []models.FactDimensionValue) error
}
