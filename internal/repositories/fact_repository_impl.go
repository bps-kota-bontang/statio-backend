package repositories

import (
	"errors"
	"fmt"
	"statio/internal/models"
	"strings"

	"gorm.io/gorm"
)

type FactRepositoryImpl struct {
	db *gorm.DB
}

// FindAllByTableAndDimensionValues implements FactRepository.
func (r *FactRepositoryImpl) FindAllByTableAndDimensionValues(tableID string, dimValueIDs []string) ([]*models.Fact, error) {
	var facts []*models.Fact

	// Jika array kosong → return semua facts
	if len(dimValueIDs) == 0 {
		err := r.db.
			Where("table_id = ?", tableID).
			Find(&facts).Error

		return facts, err
	}

	// Jika array ada isinya → filter normal
	err := r.db.Model(&models.Fact{}).
		Joins("JOIN fact_dimension_values fdv ON fdv.fact_id = facts.id").
		Where("facts.table_id = ?", tableID).
		Where("fdv.dimension_value_id IN ?", dimValueIDs).
		Group("facts.id").
		Having("COUNT(fdv.id) = ?", len(dimValueIDs)).
		Find(&facts).Error

	return facts, err
}

// CountOutliersByYear implements FactRepository.
func (r *FactRepositoryImpl) CountOutliersByYear(tableID string, fromYear, toYear int) (map[int]int, error) {
	type row struct {
		Year  int
		Count int
	}

	var rows []row

	err := r.db.
		Table("facts").
		Select("year, COUNT(*) as count").
		Where("table_id = ? AND is_outlier = TRUE AND year BETWEEN ? AND ?", tableID, fromYear, toYear).
		Group("year").
		Order("year ASC").
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	result := make(map[int]int)
	for _, r := range rows {
		result[r.Year] = r.Count
	}

	// supaya tahun yang tidak muncul tetap 0
	for year := fromYear; year <= toYear; year++ {
		if _, ok := result[year]; !ok {
			result[year] = 0
		}
	}

	return result, nil
}

// CountRevisionsByYear implements FactRepository.
func (r *FactRepositoryImpl) CountRevisionsByYear(tableID string, fromYear, toYear int) (map[int]int, error) {
	type row struct {
		Year  int
		Count int
	}

	var rows []row

	err := r.db.
		Table("facts").
		Select("year, COUNT(*) as count").
		Where(`
			table_id = ?
			AND old_value IS NOT NULL
			AND value IS NOT NULL
			AND old_value != value
			AND year BETWEEN ? AND ?
		`, tableID, fromYear, toYear).
		Group("year").
		Order("year ASC").
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	result := make(map[int]int)
	for _, r := range rows {
		result[r.Year] = r.Count
	}

	// tahun yang tidak ada tetap 0
	for year := fromYear; year <= toYear; year++ {
		if _, ok := result[year]; !ok {
			result[year] = 0
		}
	}

	return result, nil
}

func NewFactRepository(db *gorm.DB) FactRepository {
	return &FactRepositoryImpl{
		db: db,
	}
}

// ResetOutliersByTable implements FactRepository.
func (f *FactRepositoryImpl) ResetOutliersByTable(tableID string) error {
	return f.db.Model(&models.Fact{}).
		Where("table_id = ?", tableID).
		Update("is_outlier", nil).Error
}

// UpdateFact implements FactRepository.
func (f *FactRepositoryImpl) UpdateFact(fact *models.Fact) error {
	return f.db.Save(fact).Error
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

// Find all facts by table
func (f *FactRepositoryImpl) FindAllByTable(tableID string) ([]*models.Fact, error) {
	var facts []*models.Fact
	if err := f.db.Preload("FactDimensionValues").Where("table_id = ?", tableID).Find(&facts).Error; err != nil {
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
// Batch update facts with transaction using raw SQL CASE
func (f *FactRepositoryImpl) UpdateFactsTx(tx *gorm.DB, facts []*models.Fact) error {
	if len(facts) == 0 {
		return nil
	}

	// Build SQL CASE statement for "value"
	caseStmt := "CASE"
	pairs := make([]string, len(facts)) // for WHERE (id, year) IN (...)

	for i, fct := range facts {
		// ensure IDs are present
		if fct.ID == "" {
			return fmt.Errorf("fact ID is empty for update index %d", i)
		}
		pairs[i] = fmt.Sprintf("('%s', %d)", fct.ID, fct.Year)

		if fct.Value != nil {
			// use full precision float formatting
			caseStmt += fmt.Sprintf(" WHEN id = '%s' AND year = %d THEN %f", fct.ID, fct.Year, *fct.Value)
		} else {
			// set NULL for NULL values
			caseStmt += fmt.Sprintf(" WHEN id = '%s' AND year = %d THEN NULL::double precision", fct.ID, fct.Year)
		}
	}

	caseStmt += " END"

	// Build WHERE clause with (id, year)
	whereClause := fmt.Sprintf("WHERE (id, year) IN (%s)", strings.Join(pairs, ", "))

	// Final query: update only value column
	query := fmt.Sprintf("UPDATE facts SET value = %s %s", caseStmt, whereClause)

	return tx.Exec(query).Error
}

func (r *FactRepositoryImpl) CountFactsByYear(tableID string, fromYear, toYear int) (map[int]int, error) {
	type result struct {
		Year  int
		Count int
	}

	rows := []result{}
	err := r.db.Table("facts").
		Select("year, COUNT(*) as count").
		Where("value IS NOT NULL AND table_id = ? AND year BETWEEN ? AND ?", tableID, fromYear, toYear).
		Group("year").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[int]int)
	for _, r := range rows {
		counts[r.Year] = r.Count
	}

	return counts, nil
}

func (r *FactRepositoryImpl) CountOutliersForTables(tableIDs []string) (map[string]int, error) {
	type row struct {
		TableID string
		Count   int
	}

	var rows []row

	err := r.db.
		Table("facts").
		Select("table_id, COUNT(*) as count").
		Where("is_outlier = TRUE AND table_id IN ?", tableIDs).
		Group("table_id").
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	result := make(map[string]int)
	for _, r := range rows {
		result[r.TableID] = r.Count
	}

	return result, nil
}

func (r *FactRepositoryImpl) CountRevisionsForTables(tableIDs []string) (map[string]int, error) {
	type row struct {
		TableID string
		Count   int
	}

	var rows []row

	err := r.db.
		Table("facts").
		Select("table_id, COUNT(*) as count").
		Where("old_value IS NOT NULL AND value IS NOT NULL AND old_value != value AND table_id IN ?", tableIDs).
		Group("table_id").
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	result := make(map[string]int)
	for _, r := range rows {
		result[r.TableID] = r.Count
	}

	return result, nil
}

func (r *FactRepositoryImpl) CountFactsByYearForTables(tableIDs []string, fromYear, toYear int) (map[string]map[int]int, error) {
	type result struct {
		TableID string
		Year    int
		Count   int
	}

	rows := []result{}
	err := r.db.Table("facts").
		Select("table_id, year, COUNT(*) as count").
		Where("value IS NOT NULL AND table_id IN ? AND year BETWEEN ? AND ?", tableIDs, fromYear, toYear).
		Group("table_id, year").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	// Buat map[tableID][year]count
	counts := make(map[string]map[int]int)
	for _, r := range rows {
		if _, ok := counts[r.TableID]; !ok {
			counts[r.TableID] = make(map[int]int)
		}
		counts[r.TableID][r.Year] = r.Count
	}

	return counts, nil
}
