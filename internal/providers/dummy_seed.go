package providers

import (
	"errors"
	"fmt"
	"statio/internal/models"

	"gorm.io/gorm"
)

// SeedDummyData creates the minimum dataset needed to exercise tables and projects.
// It is safe to run repeatedly: records are matched by their stable dummy codes/names.
func SeedDummyData(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		population, err := seedIndicator(tx, "DUMMY_POPULATION", "Population", "Total", "People")
		if err != nil {
			return err
		}
		employment, err := seedIndicator(tx, "DUMMY_EMPLOYMENT_RATE", "Employment Rate", "Percentage", "%")
		if err != nil {
			return err
		}

		populationByDistrict, err := seedTable(tx, "Dummy Population by District", population.ID)
		if err != nil {
			return err
		}
		employmentByDistrict, err := seedTable(tx, "Dummy Employment Rate by District", employment.ID)
		if err != nil {
			return err
		}
		employmentByYear, err := seedTable(tx, "Dummy Employment Rate by Year", employment.ID)
		if err != nil {
			return err
		}

		social, err := seedProject(tx, "Dummy Project - Social Statistics")
		if err != nil {
			return err
		}
		economy, err := seedProject(tx, "Dummy Project - Economy Statistics")
		if err != nil {
			return err
		}

		links := []struct {
			projectID string
			tableID   string
			page      string
			number    string
		}{
			{social.ID, populationByDistrict.ID, "1", "1.1"},
			{social.ID, employmentByDistrict.ID, "2", "1.2"},
			{economy.ID, employmentByDistrict.ID, "5", "2.1"},
			{economy.ID, employmentByYear.ID, "6", "2.2"},
		}
		for _, link := range links {
			if err := seedProjectTable(tx, link.projectID, link.tableID, link.page, link.number); err != nil {
				return err
			}
		}

		return nil
	})
}

func seedIndicator(tx *gorm.DB, code, name, measure, unit string) (*models.Indicator, error) {
	var indicator models.Indicator
	err := tx.Unscoped().Where("code = ?", code).First(&indicator).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		codeValue := code
		unitValue := unit
		indicator = models.Indicator{Code: &codeValue, Name: name, Measure: measure, Unit: &unitValue}
	} else if err != nil {
		return nil, err
	}

	indicator.Name = name
	indicator.Measure = measure
	unitValue := unit
	indicator.Unit = &unitValue
	indicator.DeletedAt = gorm.DeletedAt{}
	if err := tx.Unscoped().Save(&indicator).Error; err != nil {
		return nil, fmt.Errorf("seed indicator %q: %w", code, err)
	}
	return &indicator, nil
}

func seedTable(tx *gorm.DB, name, indicatorID string) (*models.Table, error) {
	var table models.Table
	err := tx.Unscoped().Where("name = ?", name).First(&table).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		table = models.Table{Name: name}
	} else if err != nil {
		return nil, err
	}

	table.Name = name
	table.IndicatorID = indicatorID
	table.Direction = 1
	table.Status = "draft"
	table.IsShow = true
	table.IsIntegrated = false
	table.IsAggregated = false
	table.DeletedAt = gorm.DeletedAt{}
	if err := tx.Unscoped().Save(&table).Error; err != nil {
		return nil, fmt.Errorf("seed table %q: %w", name, err)
	}
	return &table, nil
}

func seedProject(tx *gorm.DB, name string) (*models.Project, error) {
	var project models.Project
	err := tx.Unscoped().Where("name = ?", name).First(&project).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		project = models.Project{Name: name}
	} else if err != nil {
		return nil, err
	}

	project.Name = name
	project.DeletedAt = gorm.DeletedAt{}
	if err := tx.Unscoped().Save(&project).Error; err != nil {
		return nil, fmt.Errorf("seed project %q: %w", name, err)
	}
	return &project, nil
}

func seedProjectTable(tx *gorm.DB, projectID, tableID, page, tableNumber string) error {
	var relation models.ProjectTable
	err := tx.Where("project_id = ? AND table_id = ?", projectID, tableID).First(&relation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		relation = models.ProjectTable{ProjectID: projectID, TableID: tableID}
	} else if err != nil {
		return err
	}

	relation.Page = page
	relation.TableNumber = tableNumber
	if err := tx.Save(&relation).Error; err != nil {
		return fmt.Errorf("seed project table link: %w", err)
	}
	return nil
}
