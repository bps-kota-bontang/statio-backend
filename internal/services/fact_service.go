package services

import (
	"errors"
	"fmt"
	"statio/internal/dto"
	"statio/internal/models"
	"statio/internal/repositories"
	"statio/utils"
	"time"

	"gorm.io/gorm"
)

type FactService struct {
	factRepo repositories.FactRepository
	db       *gorm.DB
}

func NewFactService(factRepo repositories.FactRepository, db *gorm.DB) *FactService {
	return &FactService{
		factRepo: factRepo,
		db:       db,
	}
}

func (s *FactService) SaveOrUpdateFacts(table *models.Table, payload *dto.UpdateFactRequest) error {
	if len(payload.Data) == 0 {
		return errors.New("no data to update")
	}

	startTotal := time.Now() // start total timer

	dimMap := utils.MapDimensionValues(table)

	// pastikan FactDimensionValues ter-load
	existingFacts, err := s.factRepo.FindAllByTableAndYear(table.ID, payload.Year)
	if err != nil {
		return err
	}

	existingFactMap := make(map[string]*models.Fact)
	for _, f := range existingFacts {
		if len(f.FactDimensionValues) == 0 {
			continue
		}
		dimIDs := make([]string, len(f.FactDimensionValues))
		for i, fdv := range f.FactDimensionValues {
			dimIDs[i] = fdv.DimensionValueID
		}
		key := utils.DimensionValueKeyFromIDs(dimIDs)
		existingFactMap[key] = f
	}

	var newFacts []models.Fact
	var newFDVs []models.FactDimensionValue
	var updateFacts []*models.Fact

	for _, factData := range payload.Data {
		if len(factData.Dimensions) == 0 {
			return errors.New("dimensions cannot be empty")
		}

		// cek apakah semua dimensi valid
		for _, dimID := range factData.Dimensions {
			if _, ok := dimMap[dimID]; !ok {
				return fmt.Errorf("dimension value not found: %s", dimID)
			}
		}

		key := utils.DimensionValueKeyFromIDs(factData.Dimensions)

		if fact, ok := existingFactMap[key]; ok {
			fact.Value = factData.Value
			updateFacts = append(updateFacts, fact)
		} else {
			newFact := models.Fact{
				TableID: table.ID,
				Year:    payload.Year,
				Value:   factData.Value,
			}
			newFacts = append(newFacts, newFact)
			for _, dimID := range factData.Dimensions {
				newFDVs = append(newFDVs, models.FactDimensionValue{
					DimensionValueID: dimID,
				})
			}
		}
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// update existing facts
	if len(updateFacts) > 0 {
		startUpdate := time.Now()
		if err := s.factRepo.UpdateFactsTx(tx, updateFacts); err != nil {
			tx.Rollback()
			return err
		}
		fmt.Printf("UpdateFacts took: %v seconds\n", time.Since(startUpdate).Seconds())
	}

	// insert new facts
	if len(newFacts) > 0 {
		startInsert := time.Now()
		if err := s.factRepo.CreateFactsTx(tx, newFacts); err != nil {
			tx.Rollback()
			return err
		}
		fmt.Printf("CreateFacts took: %v seconds\n", time.Since(startInsert).Seconds())

		// assign FactID ke FactDimensionValues
		fdvIdx := 0
		for i, f := range newFacts {
			dimsCount := len(payload.Data[i].Dimensions)
			for j := 0; j < dimsCount; j++ {
				newFDVs[fdvIdx].FactID = f.ID
				fdvIdx++
			}
		}
	}

	if len(newFDVs) > 0 {
		startFDV := time.Now()
		if err := s.factRepo.CreateFactDimensionValuesTx(tx, newFDVs); err != nil {
			tx.Rollback()
			return err
		}
		fmt.Printf("CreateFactDimensionValues took: %v seconds\n", time.Since(startFDV).Seconds())
	}

	fmt.Printf("Total SaveOrUpdateFacts took: %v seconds\n", time.Since(startTotal).Seconds())
	return tx.Commit().Error
}
