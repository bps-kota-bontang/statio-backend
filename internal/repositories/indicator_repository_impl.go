package repositories

import (
	"statio/internal/models"

	"gorm.io/gorm"
)

type IndicatorRepositoryImpl struct {
	db *gorm.DB
}

// FindAllNames implements IndicatorRepository.
func (i *IndicatorRepositoryImpl) FindAllNames() ([]*models.Indicator, error) {
	var names []*models.Indicator
	if err := i.db.Select("name").Order("name ASC").Find(&names).Error; err != nil {
		return nil, err
	}
	return names, nil
}

// FindAllMeasures implements IndicatorRepository.
func (i *IndicatorRepositoryImpl) FindAllMeasures() ([]*models.Indicator, error) {
	var measures []*models.Indicator
	if err := i.db.Select("measure").Distinct("measure").Order("measure ASC").Find(&measures).Error; err != nil {
		return nil, err
	}
	return measures, nil
}

// FindAllUnits implements IndicatorRepository.
func (i *IndicatorRepositoryImpl) FindAllUnits() ([]*models.Indicator, error) {
	var units []*models.Indicator
	if err := i.db.Select("unit").Distinct("unit").Order("unit ASC").Where("unit IS NOT NULL").Find(&units).Error; err != nil {
		return nil, err
	}
	return units, nil
}

// FindAll implements IndicatorRepository.
func (i *IndicatorRepositoryImpl) FindAll() ([]*models.Indicator, error) {
	var indicators []*models.Indicator
	if err := i.db.Find(&indicators).Error; err != nil {
		return nil, err
	}
	return indicators, nil
}

// Create implements IndicatorRepository.
func (i *IndicatorRepositoryImpl) Create(indicator *models.Indicator) error {
	if err := i.db.Create(indicator).Error; err != nil {
		return err
	}
	return nil
}

// Delete implements IndicatorRepository.
func (i *IndicatorRepositoryImpl) Delete(id string) error {
	if err := i.db.Delete(&models.Indicator{}, id).Error; err != nil {
		return err
	}
	return nil
}

// Update implements IndicatorRepository.
func (i *IndicatorRepositoryImpl) Update(indicator *models.Indicator) error {
	if err := i.db.Save(indicator).Error; err != nil {
		return err
	}
	return nil
}

// FindByID implements IndicatorRepository.
func (i *IndicatorRepositoryImpl) FindByID(id string) (*models.Indicator, error) {
	var indicator models.Indicator
	if err := i.db.First(&indicator, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &indicator, nil
}

func (r *IndicatorRepositoryImpl) Count(search string, filters map[string][]string, total *int64) error {
	query := r.db.Model(&models.Indicator{})

	if search != "" {
		like := "%" + search + "%"
		query = query.Where(
			r.db.
				Where("name ILIKE ?", like).
				Or("measure ILIKE ?", like).
				Or("unit ILIKE ?", like),
		)
	}

	// Filter per kolom
	for col, values := range filters {
		if len(values) > 0 {

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
	}

	return query.Count(total).Error
}

func (r *IndicatorRepositoryImpl) FindPaginated(
	search string,
	limit, offset int,
	sortBy, sortOrder string,
	filters map[string][]string,
) ([]*models.Indicator, error) {

	var indicators []*models.Indicator
	query := r.db.Model(&models.Indicator{}).Limit(limit).Offset(offset)

	if search != "" {
		like := "%" + search + "%"
		query = query.Where(
			r.db.
				Where("name ILIKE ?", like).
				Or("measure ILIKE ?", like).
				Or("unit ILIKE ?", like),
		)
	}

	// Filter per kolom
	for col, values := range filters {
		if len(values) > 0 {

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
	}

	// Sorting
	validSortFields := map[string]string{
		"no":      "created_at",
		"name":    "name",
		"measure": "measure",
		"unit":    "unit",
	}
	field, ok := validSortFields[sortBy]
	if !ok {
		field = "created_at"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}
	query = query.Order(field + " " + sortOrder)

	if err := query.Find(&indicators).Error; err != nil {
		return nil, err
	}
	return indicators, nil
}

func NewIndicatorRepository(db *gorm.DB) IndicatorRepository {
	return &IndicatorRepositoryImpl{db: db}
}
