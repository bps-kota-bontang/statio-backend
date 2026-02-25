package repositories

import (
	"fmt"
	"slices"
	"statio/internal/dto"
	"statio/internal/models"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type TableRepositoryImpl struct {
	db *gorm.DB
}

// Delete implements [TableRepository].
func (r *TableRepositoryImpl) Delete(tableID string) error {
	var table models.Table
	if err := r.db.Where("id = ?", tableID).First(&table).Error; err != nil {
		return err
	}

	return r.db.Delete(&table).Error
}

// FindTablesBase implements [TableRepository].
func (r *TableRepositoryImpl) FindTablesBase(filter *dto.FilterTablesRequest) ([]*models.Table, error) {
	var tables []*models.Table

	query := r.db.Model(&models.Table{})

	if filter.OrganizationID != nil {
		query = query.Where("organization_id = ?", *filter.OrganizationID)
	}

	if len(filter.TableIDs) > 0 {
		query = query.Where("id IN ?", filter.TableIDs)
	}

	if err := query.Preload("Organization").Find(&tables).Error; err != nil {
		return nil, err
	}

	return tables, nil
}

// FindByIDs implements TableRepository.
func (r *TableRepositoryImpl) FindByIDs(tableIDs []string) ([]*models.Table, error) {
	var tables []*models.Table
	if err := r.db.Preload("Indicator").Preload("Dimensions.Dimension").Preload("Organization").
		Where("id IN ?", tableIDs).
		Find(&tables).Error; err != nil {
		return nil, err
	}
	return tables, nil
}

// UpdateLabels implements TableRepository.
func (r *TableRepositoryImpl) UpdateLabels(tableID string, labels []string) error {
	return r.db.Model(&models.Table{}).
		Where("id = ?", tableID).
		Update("labels", pq.Array(labels)).
		Error
}

// FindAllLabels implements TableRepository.
func (r *TableRepositoryImpl) FindAllLabels() ([]*string, error) {
	var labels []*string

	// Jalankan raw SQL untuk ambil semua label unik
	if err := r.db.
		Model(&models.Table{}).
		Select("DISTINCT UNNEST(labels)").
		Find(&labels).Error; err != nil {
		return nil, err
	}

	return labels, nil
}

// UpdateLabelBulk implements TableRepository.
func (r *TableRepositoryImpl) AddLabelsBulk(labels []string, tableIDs []string) error {
	return r.db.
		Model(&models.Table{}).
		Where("id IN ?", tableIDs).
		Update("labels", gorm.Expr(`
			ARRAY(
				SELECT DISTINCT UNNEST(
					array_cat(COALESCE(labels, '{}'), ?)
				)
			)
		`, pq.Array(labels))).
		Error
}

// UpdateOrganizationBulk implements TableRepository.
func (j *TableRepositoryImpl) UpdateOrganizationBulk(organizationID string, tableIDs []string) error {
	return j.db.Model(&models.Table{}).
		Where("id IN ?", tableIDs).
		Update("organization_id", organizationID).Error
}

// Update implements TableRepository.
func (j *TableRepositoryImpl) Update(table *models.Table) error {
	return j.db.Save(table).Error
}

// FindBaseByID implements TableRepository.
func (j *TableRepositoryImpl) FindBaseByID(id string) (*models.Table, error) {
	var table *models.Table
	if err := j.db.
		Where("id = ?", id).
		First(&table).Error; err != nil {
		return nil, err
	}
	return table, nil
}

// Update implements TableRepository.
func (r *TableRepositoryImpl) UpdateWithRelations(table *models.Table, dimensionIDs []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update field utama
		if err := tx.Save(table).Error; err != nil {
			return err
		}

		// Jika field dimension_ids dikirim, baru proses relasi
		if dimensionIDs != nil {
			// Hapus semua relasi lama
			if err := tx.Where("table_id = ?", table.ID).Delete(&models.TableDimension{}).Error; err != nil {
				return err
			}

			// Tambahkan relasi baru (jika ada)
			if len(dimensionIDs) > 0 {
				newDimensions := make([]models.TableDimension, len(dimensionIDs))
				for i, dimID := range dimensionIDs {
					newDimensions[i] = models.TableDimension{
						TableID:     table.ID,
						DimensionID: dimID,
						Order:       i,
					}
				}
				if err := tx.Create(&newDimensions).Error; err != nil {
					return err
				}
			}

			// Update direction berdasarkan jumlah dimensi baru
			table.Direction = len(dimensionIDs)
			if err := tx.Model(table).Update("direction", table.Direction).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// CountDimensionsByTableID implements TableRepository.
func (j *TableRepositoryImpl) CountDimensionsByTableID(tableID string) (*int64, error) {
	var count int64
	err := j.db.Model(&models.TableDimension{}).
		Where("table_id = ?", tableID).
		Count(&count).Error
	if err != nil {
		return nil, err
	}

	return &count, nil
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
		case "labels":
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
				query = query.Where("?::text[] && labels OR labels IS NULL OR labels = '{}'::text[]", pq.Array(realValues))
			} else if hasNull {
				query = query.Where("labels IS NULL OR labels = '{}'::text[]")
			} else {
				query = query.Where("?::text[] && labels", pq.Array(realValues))
			}
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
				subQuery = subQuery.Where("d.name IN ?", realValues)
				query = query.Where("tables.id IN (?) OR NOT EXISTS (SELECT 1 FROM table_dimensions td2 WHERE td2.table_id = tables.id)", subQuery)
			} else if hasNull {
				query = query.Where("NOT EXISTS (SELECT 1 FROM table_dimensions td WHERE td.table_id = tables.id)")
			} else {
				subQuery = subQuery.Where("d.name IN ?", realValues)
				query = query.Where("tables.id IN (?)", subQuery)
			}

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

		case "organization_id":
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
				query = query.Where("tables.organization_id IN ? OR tables.organization_id IS NULL", realValues)
			} else if hasNull {
				query = query.Where("tables.organization_id IS NULL")
			} else {
				query = query.Where("tables.organization_id IN ?", realValues)
			}
		case "status":
			// status filter is lightweight, keep it
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
				query = query.Where("tables.status IN ? OR tables.status IS NULL", realValues)
			} else if hasNull {
				query = query.Where("tables.status IS NULL")
			} else {
				query = query.Where("tables.status IN ?", realValues)
			}
			// ignore missing_facts here (we handle in service)
		case "is_aggregated":
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
				query = query.Where("tables.is_aggregated IN ? OR tables.is_aggregated IS NULL", realValues)
			} else if hasNull {
				query = query.Where("tables.is_aggregated IS NULL")
			} else {
				query = query.Where("tables.is_aggregated IN ?", realValues)
			}
		case "is_show":
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
				query = query.Where("tables.is_show IN ? OR tables.is_show IS NULL", realValues)
			} else if hasNull {
				query = query.Where("tables.is_show IS NULL")
			} else {
				query = query.Where("tables.is_show IN ?", realValues)
			}
		case "is_integrated":
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
				query = query.Where("tables.is_integrated IN ? OR tables.is_integrated IS NULL", realValues)
			} else if hasNull {
				query = query.Where("tables.is_integrated IS NULL")
			} else {
				query = query.Where("tables.is_integrated IN ?", realValues)
			}
		case "direction":
			// direction is lightweight, keep it
			intValues := make([]int, 0, len(values))
			for _, v := range values {
				switch v {
				case "0":
					intValues = append(intValues, 0)
				case "1":
					intValues = append(intValues, 1)
				case "2":
					intValues = append(intValues, 2)
				}
			}
			if len(intValues) > 0 {
				query = query.Where("tables.direction IN ?", intValues)
			}
		case "can_integrate":
			// lightweight filter to check if tables are integratable (have website table id, subject id and website link)
			canIntegrate := slices.Contains(values, "true")
			if canIntegrate {
				query = query.Where("tables.website_table_id IS NOT NULL AND tables.website_subject_id IS NOT NULL AND tables.website_link IS NOT NULL")
			} else {
				query = query.Where("tables.website_table_id IS NULL OR tables.website_subject_id IS NULL OR tables.website_link IS NULL")
			}
		}
	}

	return query.Count(total).Error
}

// FindPaginated implements TableRepository.
func (j *TableRepositoryImpl) FindPaginated(search string, limit int, offset int, sortBy string, sortOrder string, filters map[string][]string) ([]*models.Table, error) {
	var tables []*models.Table
	query := j.db.Preload("Indicator").Preload("Dimensions.Dimension.Values").Preload("Organization").
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
		case "labels":
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
				query = query.Where("?::text[] && labels OR labels IS NULL OR labels = '{}'::text[]", pq.Array(realValues))
			} else if hasNull {
				query = query.Where("labels IS NULL OR labels = '{}'::text[]")
			} else {
				query = query.Where("?::text[] && labels", pq.Array(realValues))
			}
		case "dimensions":
			subQuery := j.db.Table("table_dimensions td").
				Select("td.table_id").
				Joins("JOIN dimensions d ON d.id = td.dimension_id").
				Where("td.deleted_at IS NULL")

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
				subQuery = subQuery.Where("d.name IN ?", realValues)
				query = query.Where("tables.id IN (?) OR NOT EXISTS (SELECT 1 FROM table_dimensions td2 WHERE td2.table_id = tables.id)", subQuery)
			} else if hasNull {
				query = query.Where("NOT EXISTS (SELECT 1 FROM table_dimensions td WHERE td.table_id = tables.id)")
			} else {
				subQuery = subQuery.Where("d.name IN ?", realValues)
				query = query.Where("tables.id IN (?)", subQuery)
			}

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
		case "organization_id":
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
				query = query.Where("tables.organization_id IN ? OR tables.organization_id IS NULL", realValues)
			} else if hasNull {
				query = query.Where("tables.organization_id IS NULL")
			} else {
				query = query.Where("tables.organization_id IN ?", realValues)
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

// FindLight implements TableRepository.
func (j *TableRepositoryImpl) FindLight(search string, sortBy string, sortOrder string, filters map[string][]string) ([]*models.Table, error) {
	var results []*models.Table

	query := j.db.Model(&models.Table{}).
		Joins("LEFT JOIN indicators i ON i.id = tables.indicator_id")

	// apply same lightweight search & filters as your Count (reuse logic)
	if search != "" {
		like := "%" + search + "%"
		query = query.Where(
			j.db.Where("tables.name ILIKE ?", like).
				Or("i.name ILIKE ?", like).
				Or("i.measure ILIKE ?", like).
				Or("i.unit ILIKE ?", like),
		)
	}

	// apply filters except heavy ones? We'll reuse your same switch but keep it as-is
	for col, values := range filters {
		if len(values) == 0 {
			continue
		}
		switch col {
		case "indicator_name":
			// same logic as Count
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
		case "organization_id":
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
				query = query.Where("tables.organization_id IN ? OR tables.organization_id IS NULL", realValues)
			} else if hasNull {
				query = query.Where("tables.organization_id IS NULL")
			} else {
				query = query.Where("tables.organization_id IN ?", realValues)
			}
		case "labels":
			// labels is not super heavy, keep it
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
				query = query.Where("?::text[] && labels OR labels IS NULL OR labels = '{}'::text[]", pq.Array(realValues))
			} else if hasNull {
				query = query.Where("labels IS NULL OR labels = '{}'::text[]")
			} else {
				query = query.Where("?::text[] && labels", pq.Array(realValues))
			}
		case "dimensions":
			// dimensions filter uses subquery (not loading dimension values), ok to keep
			subQuery := j.db.Table("table_dimensions td").
				Select("td.table_id").
				Joins("JOIN dimensions d ON d.id = td.dimension_id").
				Where("td.deleted_at IS NULL")

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
				subQuery = subQuery.Where("d.name IN ?", realValues)
				query = query.Where("tables.id IN (?) OR NOT EXISTS (SELECT 1 FROM table_dimensions td2 WHERE td2.table_id = tables.id)", subQuery)
			} else if hasNull {
				query = query.Where("NOT EXISTS (SELECT 1 FROM table_dimensions td WHERE td.table_id = tables.id)")
			} else {
				subQuery = subQuery.Where("d.name IN ?", realValues)
				query = query.Where("tables.id IN (?)", subQuery)
			}
		case "status":
			// status filter is lightweight, keep it
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
				query = query.Where("tables.status IN ? OR tables.status IS NULL", realValues)
			} else if hasNull {
				query = query.Where("tables.status IS NULL")
			} else {
				query = query.Where("tables.status IN ?", realValues)
			}
			// ignore missing_facts here (we handle in service)
		case "is_aggregated":
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
				query = query.Where("tables.is_aggregated IN ? OR tables.is_aggregated IS NULL", realValues)
			} else if hasNull {
				query = query.Where("tables.is_aggregated IS NULL")
			} else {
				query = query.Where("tables.is_aggregated IN ?", realValues)
			}
		case "is_show":
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
				query = query.Where("tables.is_show IN ? OR tables.is_show IS NULL", realValues)
			} else if hasNull {
				query = query.Where("tables.is_show IS NULL")
			} else {
				query = query.Where("tables.is_show IN ?", realValues)
			}
		case "is_integrated":
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
				query = query.Where("tables.is_integrated IN ? OR tables.is_integrated IS NULL", realValues)
			} else if hasNull {
				query = query.Where("tables.is_integrated IS NULL")
			} else {
				query = query.Where("tables.is_integrated IN ?", realValues)
			}
		case "direction":
			// direction is lightweight, keep it
			intValues := make([]int, 0, len(values))
			for _, v := range values {
				switch v {
				case "0":
					intValues = append(intValues, 0)
				case "1":
					intValues = append(intValues, 1)
				case "2":
					intValues = append(intValues, 2)
				}
			}
			if len(intValues) > 0 {
				query = query.Where("tables.direction IN ?", intValues)
			}
		case "can_integrate":
			// lightweight filter to check if tables are integratable (have website table id, subject id and website link)
			canIntegrate := slices.Contains(values, "true")
			if canIntegrate {
				query = query.Where("tables.website_table_id IS NOT NULL AND tables.website_subject_id IS NOT NULL AND tables.website_link IS NOT NULL")
			} else {
				query = query.Where("tables.website_table_id IS NULL OR tables.website_subject_id IS NULL OR tables.website_link IS NULL")
			}
		}
	}

	// ordering - reuse existing mapping
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

	// select minimal columns
	if err := query.Select("tables.id, tables.name, tables.indicator_id").Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

// LoadDimensionsForTableIDs implements TableRepository.
func (j *TableRepositoryImpl) LoadDimensionsForTableIDs(ids []string) ([]*models.Table, error) {
	if len(ids) == 0 {
		return []*models.Table{}, nil
	}
	var tables []*models.Table
	err := j.db.Model(&models.Table{}).
		Where("id IN ?", ids).
		// Preload("Dimensions").
		// Preload("Dimensions.Dimension").
		Preload("Dimensions.Dimension.Values").
		Find(&tables).Error
	if err != nil {
		return nil, err
	}
	return tables, nil
}

// FindByIDsDetailed implements TableRepository.
func (j *TableRepositoryImpl) FindByIDsDetailed(ids []string) ([]*models.Table, error) {
	if len(ids) == 0 {
		return []*models.Table{}, nil
	}
	var tables []*models.Table
	err := j.db.
		Preload("Indicator").
		Preload("Dimensions", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`).Preload("Dimension.Values", func(db2 *gorm.DB) *gorm.DB {
				return db2.Order(`"order" ASC`)
			})
		}).
		//Preload("Facts"). // adjust if you need specific fact preloads
		Preload("Organization").
		Where("id IN ?", ids).
		Find(&tables).Error
	if err != nil {
		return nil, err
	}
	return tables, nil
}

// FindAll implements TableRepository.
func (j *TableRepositoryImpl) FindAll() ([]*models.Table, error) {
	var tables []*models.Table
	if err := j.db.Preload("Indicator").Preload("Dimensions.Dimension").Preload("Organization").Find(&tables).Error; err != nil {
		return nil, err
	}
	return tables, nil
}

// FindByIDForFactUpdate implements TableRepository.
func (j *TableRepositoryImpl) FindForFactUpdate(id string) (*models.Table, error) {
	var table models.Table
	if err := j.db.
		// Preload("Dimensions.Dimension.Values", func(db *gorm.DB) *gorm.DB {
		// 	return db.Order(`"order" ASC`)
		// }).
		Preload("Dimensions", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`).Preload("Dimension.Values", func(db2 *gorm.DB) *gorm.DB {
				return db2.Order(`"order" ASC`)
			})
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

// BeginTx implements TableRepository.
func (r *TableRepositoryImpl) BeginTx() *gorm.DB {
	return r.db.Begin()
}

// CreateTableDimensionWithTx implements TableRepository.
func (r *TableRepositoryImpl) CreateTableDimensionWithTx(tx *gorm.DB, td *models.TableDimension) error {
	return tx.Create(td).Error
}

// FindBySourceTableID implements TableRepository.
func (r *TableRepositoryImpl) FindBySourceTableID(sourceTableID string) (*models.Table, error) {
	var table models.Table
	if err := r.db.Preload("Dimensions").
		Where("source_table_id = ? AND is_aggregated = ?", sourceTableID, true).
		First(&table).Error; err != nil {
		return nil, err
	}
	return &table, nil
}

// FindAllBySourceTableID implements TableRepository.
func (r *TableRepositoryImpl) FindAllBySourceTableID(sourceTableID string) ([]*models.Table, error) {
	var tables []*models.Table
	if err := r.db.Preload("Dimensions").
		Where("source_table_id = ? AND is_aggregated = ?", sourceTableID, true).
		Find(&tables).Error; err != nil {
		return nil, err
	}
	return tables, nil
}

func (j *TableRepositoryImpl) FindDetailedByID(id string) (*models.Table, error) {
	var table models.Table
	if err := j.db.
		Preload("Indicator").
		Preload("Facts.FactDimensionValues.DimensionValue.Dimension").
		Preload("Organization").
		// Preload("Dimensions.Dimension.Values", func(db *gorm.DB) *gorm.DB {
		// 	return db.Order(`"order" ASC`)
		// }).
		Preload("Dimensions", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`).Preload("Dimension.Values", func(db2 *gorm.DB) *gorm.DB {
				return db2.Order(`"order" ASC`)
			})
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
		Preload("Organization").
		Preload("Facts.FactDimensionValues.DimensionValue.Dimension").
		// Preload("Dimensions.Dimension.Values", func(db *gorm.DB) *gorm.DB {
		// 	return db.Order(`"order" ASC`)
		// }).
		Preload("Dimensions", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`).Preload("Dimension.Values", func(db2 *gorm.DB) *gorm.DB {
				return db2.Order(`"order" ASC`).Preload("Parent", func(db3 *gorm.DB) *gorm.DB {
					return db3.Order(`"order" ASC`)
				})
			})
		}).
		Where("id = ?", id).
		First(&table).Error; err != nil {
		return nil, err
	}
	return &table, nil
}

func (j *TableRepositoryImpl) FindByIDAndMultiYear(id string, year []int) (*models.Table, error) {
	var table models.Table
	if err := j.db.
		Preload("Indicator").
		Preload("Facts", "year IN ?", year).
		Preload("Organization").
		Preload("Facts.FactDimensionValues.DimensionValue.Dimension").
		// Preload("Dimensions.Dimension.Values", func(db *gorm.DB) *gorm.DB {
		// 	return db.Order(`"order" ASC`)
		// }).
		Preload("Dimensions", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`).Preload("Dimension.Values", func(db2 *gorm.DB) *gorm.DB {
				return db2.Order(`"order" ASC`)
			})
		}).
		Where("id = ?", id).
		First(&table).Error; err != nil {
		return nil, err
	}
	return &table, nil
}

// SwapTableDimensions implements TableRepository.
func (j *TableRepositoryImpl) SwapTableDimensions(tableID string) error {
	return j.db.Transaction(func(tx *gorm.DB) error {
		// Load table with dimensions
		var table models.Table
		if err := tx.Preload("Dimensions", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).Where("id = ?", tableID).First(&table).Error; err != nil {
			return err
		}

		if len(table.Dimensions) < 2 {
			return fmt.Errorf("table must have at least 2 dimensions to swap")
		}

		// Get the first two dimensions
		dim1 := &table.Dimensions[0]
		dim2 := &table.Dimensions[1]

		// Swap their order values
		tempOrder := dim1.Order
		dim1.Order = dim2.Order
		dim2.Order = tempOrder

		// Update both dimensions in database
		if err := tx.Model(&models.TableDimension{}).
			Where("id = ?", dim1.ID).
			Update("order", dim1.Order).Error; err != nil {
			return fmt.Errorf("failed to update first dimension order: %w", err)
		}

		if err := tx.Model(&models.TableDimension{}).
			Where("id = ?", dim2.ID).
			Update("order", dim2.Order).Error; err != nil {
			return fmt.Errorf("failed to update second dimension order: %w", err)
		}

		return nil
	})
}
