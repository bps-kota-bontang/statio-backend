package repositories

import (
	"statio/internal/dto"
	"statio/internal/models"

	"gorm.io/gorm"
)

type ProjectRepositoryImpl struct {
	db *gorm.DB
}

func (r *ProjectRepositoryImpl) Count(search string, total *int64) error {
	query := r.db.Model(&models.Project{})
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}
	return query.Count(total).Error
}

func (r *ProjectRepositoryImpl) FindPaginated(search string, limit, offset int, sortBy, sortOrder string) ([]*models.Project, error) {
	var projects []*models.Project
	query := r.db.Model(&models.Project{}).Preload("Tables").Limit(limit).Offset(offset)
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	field := map[string]string{"no": "created_at", "name": "name"}[sortBy]
	if field == "" {
		field = "created_at"
	}
	if sortOrder != "desc" {
		sortOrder = "asc"
	}

	if err := query.Order(field + " " + sortOrder).Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *ProjectRepositoryImpl) FindByID(id string) (*models.Project, error) {
	var project models.Project
	if err := r.db.Preload("Tables.Table").First(&project, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepositoryImpl) Create(project *models.Project) error {
	return r.db.Create(project).Error
}

func (r *ProjectRepositoryImpl) Update(project *models.Project) error {
	return r.db.Save(project).Error
}

func (r *ProjectRepositoryImpl) Delete(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var project models.Project
		if err := tx.First(&project, "id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&models.ProjectTable{}).Error; err != nil {
			return err
		}
		return tx.Delete(&project).Error
	})
}

func (r *ProjectRepositoryImpl) AddTable(projectID string, input *dto.ProjectTableRequest) error {
	if err := r.db.First(&models.Project{}, "id = ?", projectID).Error; err != nil {
		return err
	}
	if err := r.db.First(&models.Table{}, "id = ?", input.TableID).Error; err != nil {
		return err
	}

	var existing models.ProjectTable
	if err := r.db.Where("project_id = ? AND table_id = ?", projectID, input.TableID).First(&existing).Error; err == nil {
		return ErrProjectTableExists
	} else if err != gorm.ErrRecordNotFound {
		return err
	}

	return r.db.Create(&models.ProjectTable{
		ProjectID:   projectID,
		TableID:     input.TableID,
		Page:        input.Page,
		TableNumber: input.TableNumber,
	}).Error
}

func (r *ProjectRepositoryImpl) UpdateTable(projectID, tableID string, input *dto.ProjectTableRequest) error {
	result := r.db.Model(&models.ProjectTable{}).
		Where("project_id = ? AND table_id = ?", projectID, tableID).
		Updates(map[string]any{
			"page":         input.Page,
			"table_number": input.TableNumber,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *ProjectRepositoryImpl) RemoveTable(projectID, tableID string) error {
	result := r.db.Where("project_id = ? AND table_id = ?", projectID, tableID).Delete(&models.ProjectTable{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &ProjectRepositoryImpl{db: db}
}
