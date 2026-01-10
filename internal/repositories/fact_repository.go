package repositories

import (
	"statio/internal/models"

	"gorm.io/gorm"
)

type FactRepository interface {
	FindFactByDimensionValues(tableID string, year int, dimValueIDs []string) (*models.Fact, error)
	FindAllByTableAndYear(tableID string, year int) ([]*models.Fact, error)
	FindAllByTable(tableID string) ([]*models.Fact, error)
	FindAllByTableAndDimensionValues(tableID string, dimValueIDs []string) ([]*models.Fact, error)
	CreateFactsTx(tx *gorm.DB, facts []models.Fact) error
	UpdateFactsTx(tx *gorm.DB, facts []*models.Fact) error
	CreateFactDimensionValuesTx(tx *gorm.DB, fdvs []models.FactDimensionValue) error
	CountFactsByYear(tableID string, fromYear, toYear int) (map[int]int, error)
	CountOutliersByYear(tableID string, fromYear, toYear int) (map[int]int, error)
	CountRevisionsByYear(tableID string, fromYear, toYear int) (map[int]int, error)
	CountFactsByYearForTables(tableIDs []string, fromYear, toYear int) (map[string]map[int]int, error)
	CountOutliersForTables(tableIDs []string) (map[string]int, error)
	CountRevisionsForTables(tableIDs []string) (map[string]int, error)
	UpdateFact(fact *models.Fact) error
	ResetOutliersByTable(tableID string) error
	CommitByTable(tableID string) error
	FindFactDimensionValuesByFactIDs(factIDs []string) ([]models.FactDimensionValue, error)
	FindFactsByTableID(tableID string) ([]models.Fact, error)
	DeleteFactDimensionValuesByFactIDs(factIDs []string) error
	DeleteFactsByIDs(factIDs []string) error
	CreateFactWithTx(tx *gorm.DB, fact *models.Fact) error
	CreateFactDimensionValueWithTx(tx *gorm.DB, fdv *models.FactDimensionValue) error
	BeginTx() *gorm.DB
	FindDimensionValueByID(id string) (*models.DimensionValue, error)
}
