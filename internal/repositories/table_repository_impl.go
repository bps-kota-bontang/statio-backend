package repositories

import (
	"statio/internal/models"

	"gorm.io/gorm"
)

type TableRepositoryImpl struct {
	db *gorm.DB
}

// CreateWithTx implements TableRepository.
func (j *TableRepositoryImpl) CreateWithTx(tx *gorm.DB, table *models.Table) error {
	return tx.Create(table).Error
}

// Count implements TableRepository.
func (j *TableRepositoryImpl) Count(search string, filters map[string][]string, total *int64) error {
	query := j.db.Model(&models.Table{}).
		Joins("LEFT JOIN indicators i ON i.id = tables.indicator_id")

	// Search across table and indicator fields
	if search != "" {
		like := "%" + search + "%"
		query = query.Where(
			j.db.Where("tables.name ILIKE ?", like).
				Or("i.name ILIKE ?", like).
				Or("i.measure ILIKE ?", like).
				Or("i.unit ILIKE ?", like),
		)
	}

	// Filter per kolom
	for col, values := range filters {
		if len(values) == 0 {
			continue
		}

		switch col {
		case "dimensions":
			subQuery := j.db.Table("table_dimensions td").
				Select("td.table_id").
				Joins("JOIN dimensions d ON d.id = td.dimension_id")

			hasNull := false
			realValues := make([]string, 0, len(values))
			for _, v := range values {
				if v == "__NULL__" {
					hasNull = true
				} else {
					realValues = append(realValues, v)
				}
			}

			if hasNull && len(realValues) > 0 {
				subQuery = subQuery.Where("d.name IN ? OR d.name IS NULL OR d.name = ''", realValues)
			} else if hasNull {
				subQuery = subQuery.Where("d.name IS NULL OR d.name = ''")
			} else {
				subQuery = subQuery.Where("d.name IN ?", realValues)
			}

			query = query.Where("tables.id IN (?)", subQuery)

		case "indicator_name":
			hasNull := false
			realValues := make([]string, 0, len(values))
			for _, v := range values {
				if v == "__NULL__" {
					hasNull = true
				} else {
					realValues = append(realValues, v)
				}
			}
			if hasNull && len(realValues) > 0 {
				query = query.Where("i.name IN ? OR i.name IS NULL OR i.name = ''", realValues)
			} else if hasNull {
				query = query.Where("i.name IS NULL OR i.name = ''")
			} else {
				query = query.Where("i.name IN ?", realValues)
			}

		case "indicator_measure":
			hasNull := false
			realValues := make([]string, 0, len(values))
			for _, v := range values {
				if v == "__NULL__" {
					hasNull = true
				} else {
					realValues = append(realValues, v)
				}
			}
			if hasNull && len(realValues) > 0 {
				query = query.Where("i.measure IN ? OR i.measure IS NULL OR i.measure = ''", realValues)
			} else if hasNull {
				query = query.Where("i.measure IS NULL OR i.measure = ''")
			} else {
				query = query.Where("i.measure IN ?", realValues)
			}

		case "indicator_unit":
			hasNull := false
			realValues := make([]string, 0, len(values))
			for _, v := range values {
				if v == "__NULL__" {
					hasNull = true
				} else {
					realValues = append(realValues, v)
				}
			}
			if hasNull && len(realValues) > 0 {
				query = query.Where("i.unit IN ? OR i.unit IS NULL OR i.unit = ''", realValues)
			} else if hasNull {
				query = query.Where("i.unit IS NULL OR i.unit = ''")
			} else {
				query = query.Where("i.unit IN ?", realValues)
			}
		}
	}

	return query.Count(total).Error
}

// FindPaginated implements TableRepository.
func (j *TableRepositoryImpl) FindPaginated(search string, limit int, offset int, sortBy string, sortOrder string, filters map[string][]string) ([]*models.Table, error) {
	var tables []*models.Table
	query := j.db.Preload("Indicator").Preload("Dimensions.Dimension").
		Model(&models.Table{}).
		Joins("LEFT JOIN indicators i ON i.id = tables.indicator_id").
		Limit(limit).Offset(offset)

	// Search across table and indicator fields
	if search != "" {
		like := "%" + search + "%"
		query = query.Where(
			j.db.Where("tables.name ILIKE ?", like).
				Or("i.name ILIKE ?", like).
				Or("i.measure ILIKE ?", like).
				Or("i.unit ILIKE ?", like),
		)
	}

	// Filter per kolom
	for col, values := range filters {
		if len(values) == 0 {
			continue
		}

		switch col {
		case "dimensions":
			subQuery := j.db.Table("table_dimensions td").
				Select("td.table_id").
				Joins("JOIN dimensions d ON d.id = td.dimension_id")

			hasNull := false
			realValues := make([]string, 0, len(values))
			for _, v := range values {
				if v == "__NULL__" {
					hasNull = true
				} else {
					realValues = append(realValues, v)
				}
			}

			if hasNull && len(realValues) > 0 {
				subQuery = subQuery.Where("d.name IN ? OR d.name IS NULL OR d.name = ''", realValues)
			} else if hasNull {
				subQuery = subQuery.Where("d.name IS NULL OR d.name = ''")
			} else {
				subQuery = subQuery.Where("d.name IN ?", realValues)
			}

			query = query.Where("tables.id IN (?)", subQuery)

		case "indicator_name":
			hasNull := false
			realValues := make([]string, 0, len(values))
			for _, v := range values {
				if v == "__NULL__" {
					hasNull = true
				} else {
					realValues = append(realValues, v)
				}
			}
			if hasNull && len(realValues) > 0 {
				query = query.Where("i.name IN ? OR i.name IS NULL OR i.name = ''", realValues)
			} else if hasNull {
				query = query.Where("i.name IS NULL OR i.name = ''")
			} else {
				query = query.Where("i.name IN ?", realValues)
			}

		case "indicator_measure":
			hasNull := false
			realValues := make([]string, 0, len(values))
			for _, v := range values {
				if v == "__NULL__" {
					hasNull = true
				} else {
					realValues = append(realValues, v)
				}
			}
			if hasNull && len(realValues) > 0 {
				query = query.Where("i.measure IN ? OR i.measure IS NULL OR i.measure = ''", realValues)
			} else if hasNull {
				query = query.Where("i.measure IS NULL OR i.measure = ''")
			} else {
				query = query.Where("i.measure IN ?", realValues)
			}

		case "indicator_unit":
			hasNull := false
			realValues := make([]string, 0, len(values))
			for _, v := range values {
				if v == "__NULL__" {
					hasNull = true
				} else {
					realValues = append(realValues, v)
				}
			}
			if hasNull && len(realValues) > 0 {
				query = query.Where("i.unit IN ? OR i.unit IS NULL OR i.unit = ''", realValues)
			} else if hasNull {
				query = query.Where("i.unit IS NULL OR i.unit = ''")
			} else {
				query = query.Where("i.unit IN ?", realValues)
			}
		}
	}

	// Sorting
	validSortFields := map[string]string{
		"no":   "tables.created_at",
		"name": "tables.name",
	}
	field, ok := validSortFields[sortBy]
	if !ok {
		field = "tables.created_at"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}
	query = query.Order(field + " " + sortOrder)

	if err := query.Find(&tables).Error; err != nil {
		return nil, err
	}
	return tables, nil
}

// FindAll implements TableRepository.
func (j *TableRepositoryImpl) FindAll() ([]*models.Table, error) {
	var tables []*models.Table
	if err := j.db.Preload("Indicator").Preload("Dimensions.Dimension").Find(&tables).Error; err != nil {
		return nil, err
	}
	return tables, nil
}

// FindByIDForFactUpdate implements TableRepository.
func (j *TableRepositoryImpl) FindByIDForFactUpdate(id string) (*models.Table, error) {
	var table models.Table
	if err := j.db.
		Preload("Dimensions.Dimension.Values", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Where("id = ?", id).
		First(&table).Error; err != nil {
		return nil, err
	}
	return &table, nil
}

func NewTableRepository(db *gorm.DB) TableRepository {
	return &TableRepositoryImpl{
		db: db,
	}
}

func (j *TableRepositoryImpl) FindByID(id string) (*models.Table, error) {
	var table models.Table
	if err := j.db.
		Preload("Indicator").
		Preload("Facts.FactDimensionValues.DimensionValue.Dimension").
		Preload("Dimensions.Dimension.Values", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Where("id = ?", id).
		First(&table).Error; err != nil {
		return nil, err
	}
	return &table, nil
}

func (j *TableRepositoryImpl) FindByIDAndYear(id string, year int) (*models.Table, error) {
	var table models.Table
	if err := j.db.
		Preload("Indicator").
		Preload("Facts", "year = ?", year).
		Preload("Facts.FactDimensionValues.DimensionValue.Dimension").
		Preload("Dimensions.Dimension.Values", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Where("id = ?", id).
		First(&table).Error; err != nil {
		return nil, err
	}
	return &table, nil
}
