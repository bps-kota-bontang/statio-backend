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
	existingFacts, err := s.factRepo.FindAllByTable(table.ID)
	if err != nil {
		return err
	}

	existingFactMap := make(map[string]*models.Fact)
	for _, f := range existingFacts {
		dimIDs := make([]string, len(f.FactDimensionValues))
		for i, fdv := range f.FactDimensionValues {
			dimIDs[i] = fdv.DimensionValueID
		}

		// sertakan year agar key unik per tahun
		key := utils.DimensionValueYearKey(f.Year, dimIDs)
		existingFactMap[key] = f
	}

	var newFacts []models.Fact
	var newFDVs []models.FactDimensionValue
	var updateFacts []*models.Fact

	// track dimensions per new fact so we can assign FDVs correctly after creating facts
	var dimsForNewFact [][]string

	for _, factData := range payload.Data {
		// cek apakah semua dimensi valid
		for _, dimID := range factData.Dimensions {
			if _, ok := dimMap[dimID]; !ok {
				return fmt.Errorf("dimension value not found: %s", dimID)
			}
		}

		key := utils.DimensionValueYearKey(factData.Year, factData.Dimensions)

		if fact, ok := existingFactMap[key]; ok {
			// update existing fact (we keep pointer to update via batch)
			fact.Value = factData.Value
			fact.Year = factData.Year
			updateFacts = append(updateFacts, fact)
		} else {
			// prepare new fact and its FDVs
			newFact := models.Fact{
				TableID: table.ID,
				Year:    factData.Year,
				Value:   factData.Value,
			}
			newFacts = append(newFacts, newFact)
			// append the dimensions for this new fact so we know how many FDVs belong to it
			dimsForNewFact = append(dimsForNewFact, append([]string{}, factData.Dimensions...))

			// create FDV entries without FactID for now; we'll set FactID after insert
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
			panic(r) // re-panic after rollback so caller knows something bad happened
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

		// assign FactID ke FactDimensionValues using dimsForNewFact to know counts per new fact
		fdvIdx := 0
		for i := range newFacts {
			// some ORMs populate IDs after Create; ensure we read the ID from newFacts[i]
			factID := newFacts[i].ID
			if factID == "" {
				tx.Rollback()
				return fmt.Errorf("fact ID is empty after insert for newFacts index %d", i)
			}

			dimsCount := len(dimsForNewFact[i])
			for range dimsCount {
				// guard index
				if fdvIdx >= len(newFDVs) {
					tx.Rollback()
					return fmt.Errorf("internal error: fdv index out of range")
				}
				newFDVs[fdvIdx].FactID = factID
				fdvIdx++
			}
		}
	}

	// insert FDVs
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

func (s *FactService) GetMissingFactsForTable(table *models.Table, fromYear, toYear int) (*dto.MissingFactsResponse, error) {
	// Hitung jumlah kombinasi dimensi
	expectedPerYear := 1
	for _, td := range table.Dimensions {
		if td.Dimension != nil {
			expectedPerYear *= len(td.Dimension.Values)
		}
	}

	// Ambil jumlah faktual per tahun dari DB
	filledCounts, err := s.factRepo.CountFactsByYear(table.ID, fromYear, toYear)
	if err != nil {
		return nil, fmt.Errorf("failed to count facts: %w", err)
	}

	// Bangun response model
	data := make([]dto.DataMissingFact, 0, toYear-fromYear+1)
	totalExpected := 0
	totalFilled := 0
	totalMissing := 0

	for year := fromYear; year <= toYear; year++ {
		filled := filledCounts[year]
		missing := max(expectedPerYear-filled, 0)
		data = append(data, dto.DataMissingFact{
			Year:     year,
			Expected: expectedPerYear,
			Filled:   filled,
			Missing:  missing,
		})

		totalExpected += expectedPerYear
		totalFilled += filled
		totalMissing += missing
	}

	return &dto.MissingFactsResponse{
		TableID:  table.ID,
		FromYear: fromYear,
		ToYear:   toYear,
		Summary: dto.SummaryMissingFacts{
			ExpectedPerYear: expectedPerYear,
			TotalExpected:   totalExpected,
			TotalFilled:     totalFilled,
			TotalMissing:    totalMissing,
		},
		Data: data,
	}, nil
}
