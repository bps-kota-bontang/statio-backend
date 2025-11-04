package repositories

import (
	"errors"
	"fmt"
	"statio/internal/models"

	"gorm.io/gorm"
)

type FactRepositoryImpl struct {
	db *gorm.DB
}

func NewFactRepository(db *gorm.DB) FactRepository {
	return &FactRepositoryImpl{
		db: db,
	}
}

// Find single fact by dimensions
func (f *FactRepositoryImpl) FindFactByDimensionValues(tableID string, year int, dimValueIDs []string) (*models.Fact, error) {
	if len(dimValueIDs) == 0 {
		return nil, errors.New("dimValueIDs cannot be empty")
	}

	var fact models.Fact
	err := f.db.Model(&models.Fact{}).
		Joins("JOIN fact_dimension_values fdv ON fdv.fact_id = facts.id").
		Where("facts.table_id = ? AND facts.year = ?", tableID, year).
		Where("fdv.dimension_value_id IN ?", dimValueIDs).
		Group("facts.id").
		Having("COUNT(fdv.id) = ?", len(dimValueIDs)).
		Limit(1).
		Find(&fact).Error

	if err != nil {
		return nil, err
	}
	if fact.ID == "" {
		return nil, nil
	}
	return &fact, nil
}

// Find all facts by table + year
func (f *FactRepositoryImpl) FindAllByTableAndYear(tableID string, year int) ([]*models.Fact, error) {
	var facts []*models.Fact
	if err := f.db.Preload("FactDimensionValues").Where("table_id = ? AND year = ?", tableID, year).Find(&facts).Error; err != nil {
		return nil, err
	}
	return facts, nil
}

// Batch insert facts with transaction
func (f *FactRepositoryImpl) CreateFactsTx(tx *gorm.DB, facts []models.Fact) error {
	if len(facts) == 0 {
		return nil
	}
	return tx.Create(&facts).Error
}

// Batch insert FactDimensionValues with transaction
func (f *FactRepositoryImpl) CreateFactDimensionValuesTx(tx *gorm.DB, fdvs []models.FactDimensionValue) error {
	if len(fdvs) == 0 {
		return nil
	}
	return tx.Create(&fdvs).Error
}

// Batch update facts with transaction using raw SQL CASE
func (f *FactRepositoryImpl) UpdateFactsTx(tx *gorm.DB, facts []*models.Fact) error {
	if len(facts) == 0 {
		return nil
	}

	// Build SQL CASE statement
	ids := make([]string, len(facts))
	caseStmt := "CASE id "
	for i, fct := range facts {
		ids[i] = fct.ID
		if fct.Value != nil {
			caseStmt += fmt.Sprintf("WHEN '%s' THEN %f ", fct.ID, *fct.Value)
		} else {
			caseStmt += fmt.Sprintf("WHEN '%s' THEN NULL::double precision ", fct.ID)
		}
	}
	caseStmt += "END"

	query := fmt.Sprintf("UPDATE facts SET value = %s WHERE id IN ?", caseStmt)

	return tx.Exec(query, ids).Error
}
