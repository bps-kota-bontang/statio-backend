package repositories

import (
	"statio/internal/models"

	"gorm.io/gorm"
)

type DimensionRepositoryImpl struct {
	db *gorm.DB
}

// AreValidIDs implements DimensionRepository.
func (r *DimensionRepositoryImpl) AreValidIDs(ids []string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Dimension{}).
		Where("id IN ?", ids).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count == int64(len(ids)), nil
}

// FindAllNames implements DimensionRepository.
func (i *DimensionRepositoryImpl) FindAllNames() ([]*models.Dimension, error) {
	var dimensions []*models.Dimension
	if err := i.db.
		Select("name").
		Order("name ASC").
		Find(&dimensions).Error; err != nil {
		return nil, err
	}
	return dimensions, nil
}

// FindAll implements DimensionRepository.
func (i *DimensionRepositoryImpl) FindAll() ([]*models.Dimension, error) {
	var dimensions []*models.Dimension
	if err := i.db.Find(&dimensions).Error; err != nil {
		return nil, err
	}
	return dimensions, nil
}

// Create implements DimensionRepository.
func (i *DimensionRepositoryImpl) Create(dimension *models.Dimension) error {
	if err := i.db.Create(dimension).Error; err != nil {
		return err
	}
	return nil
}

// Delete implements DimensionRepository.
func (i *DimensionRepositoryImpl) Delete(id string) error {
	if err := i.db.Delete(&models.Dimension{}, id).Error; err != nil {
		return err
	}
	return nil
}

// Update implements DimensionRepository.
func (i *DimensionRepositoryImpl) Update(dimension *models.Dimension) error {
	tx := i.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Ambil semua ID value lama dari DB
	var existingIDs []string
	if err := tx.Model(&models.DimensionValue{}).
		Where("dimension_id = ?", dimension.ID).
		Pluck("id", &existingIDs).Error; err != nil {
		tx.Rollback()
		return err
	}

	existingMap := make(map[string]bool)
	for _, id := range existingIDs {
		existingMap[id] = true
	}

	// ID yang masih dipertahankan (dari payload)
	retainedIDs := make(map[string]bool)
	for _, v := range dimension.Values {
		if v.ID != "" {
			retainedIDs[v.ID] = true
		}
	}

	// Hapus value yang tidak dipertahankan
	var deletedIDs []string
	for _, id := range existingIDs {
		if !retainedIDs[id] {
			deletedIDs = append(deletedIDs, id)
		}
	}
	if len(deletedIDs) > 0 {
		if err := tx.Delete(&models.DimensionValue{}, deletedIDs).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Update nama dimension
	if err := tx.Model(&models.Dimension{}).
		Where("id = ?", dimension.ID).
		Update("name", dimension.Name).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Loop setiap value
	for _, v := range dimension.Values {
		if v.ID == "" {
			// Value baru → buat
			v.DimensionID = dimension.ID
			if err := tx.Create(&v).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else {
			// Value lama → update name dan order
			if err := tx.Model(&models.DimensionValue{}).
				Where("id = ?", v.ID).
				Updates(map[string]any{
					"name":  v.Name,
					"order": v.Order,
				}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

// FindByID implements DimensionRepository.
func (i *DimensionRepositoryImpl) FindByID(id string) (*models.Dimension, error) {
	var dimension models.Dimension
	if err := i.db.Preload("Values", func(db *gorm.DB) *gorm.DB {
		return db.Order(`"order" ASC`)
	}).First(&dimension, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &dimension, nil
}

func (r *DimensionRepositoryImpl) Count(search string, filters map[string][]string, total *int64) error {
	query := r.db.Model(&models.Dimension{})

	// JOIN ke DimensionValue untuk search
	if search != "" {
		like := "%" + search + "%"
		query = query.Joins("LEFT JOIN dimension_values dv ON dv.dimension_id = dimensions.id").
			Where("dimensions.name ILIKE ? OR dv.name ILIKE ?", like, like)
	}

	// Filter per kolom
	for col, values := range filters {
		if len(values) == 0 {
			continue
		}

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
			query = query.Where("("+col+" IN ? OR "+col+" IS NULL OR "+col+" = '')", realValues)
		} else if hasNull {
			query = query.Where(col + " IS NULL OR " + col + " = ''")
		} else {
			query = query.Where(col+" IN ?", realValues)
		}
	}

	// Gunakan DISTINCT supaya count tidak double karena JOIN
	return query.Distinct("dimensions.id").Count(total).Error
}

func (r *DimensionRepositoryImpl) FindPaginated(
	search string,
	limit, offset int,
	sortBy, sortOrder string,
	filters map[string][]string,
) ([]*models.Dimension, error) {

	// 1. Subquery untuk ambil ID unik dimension
	subQuery := r.db.Model(&models.Dimension{})

	if search != "" {
		like := "%" + search + "%"
		subQuery = subQuery.Joins("LEFT JOIN dimension_values dv ON dv.dimension_id = dimensions.id").
			Where("dimensions.name ILIKE ? OR dv.name ILIKE ?", like, like)
	}

	// Filter per kolom
	for col, values := range filters {
		if len(values) == 0 {
			continue
		}

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
			subQuery = subQuery.Where("("+col+" IN ? OR "+col+" IS NULL OR "+col+" = '')", realValues)
		} else if hasNull {
			subQuery = subQuery.Where(col + " IS NULL OR " + col + " = ''")
		} else {
			subQuery = subQuery.Where(col+" IN ?", realValues)
		}
	}

	// Sorting
	validSortFields := map[string]string{
		"no":   "created_at",
		"name": "name",
	}
	field, ok := validSortFields[sortBy]
	if !ok {
		field = "created_at"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}

	// DISTINCT untuk mencegah duplicate akibat JOIN
	subQuery = subQuery.Distinct("dimensions.id").Limit(limit).Offset(offset)

	// Ambil list ID
	var ids []string
	if err := subQuery.Pluck("dimensions.id", &ids).Error; err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []*models.Dimension{}, nil
	}

	// 2. Query utama dengan preload Values
	var dimensions []*models.Dimension
	if err := r.db.Preload("Values",
		func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).Where("id IN ?", ids).
		Order(field + " " + sortOrder).
		Find(&dimensions).Error; err != nil {
		return nil, err
	}

	return dimensions, nil
}

// FindByIDWithValues mengambil dimension beserta semua values, parent, dan children
func (r *DimensionRepositoryImpl) FindByIDWithValues(id string) (*models.Dimension, error) {
	var dimension models.Dimension
	if err := r.db.
		Preload("Values", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Preload("Values.Parent").
		Preload("Values.Children").
		First(&dimension, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &dimension, nil
}

// FindParentValueIDs mengambil semua parent value IDs dari dimension value yang memiliki parent
func (r *DimensionRepositoryImpl) FindParentValueIDs(dimensionID string) ([]string, error) {
	var parentIDs []string
	if err := r.db.Model(&models.DimensionValue{}).
		Where("dimension_id = ? AND parent_id IS NOT NULL", dimensionID).
		Distinct("parent_id").
		Pluck("parent_id", &parentIDs).Error; err != nil {
		return nil, err
	}
	return parentIDs, nil
}

// FindDimensionByID implements DimensionRepository - loads dimension without values.
func (r *DimensionRepositoryImpl) FindDimensionByID(id string) (*models.Dimension, error) {
	var dimension models.Dimension
	if err := r.db.Where("id = ?", id).First(&dimension).Error; err != nil {
		return nil, err
	}
	return &dimension, nil
}

func NewDimensionRepository(db *gorm.DB) DimensionRepository {
	return &DimensionRepositoryImpl{db: db}
}
