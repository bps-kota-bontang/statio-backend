package services

import (
	"errors"
	"fmt"
	"log"
	"statio/internal/dto"
	"statio/internal/mappers"
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

func (s *FactService) GetInsightFactsForTable(table *models.Table, fromYear, toYear int) (*dto.InsightFactsResponse, error) {
	// Validasi tahun
	if fromYear > toYear {
		return nil, fmt.Errorf("invalid year range: fromYear %d > toYear %d", fromYear, toYear)
	}

	// Hitung kombinasi dimensi
	expectedPerYear := 1
	for _, td := range table.Dimensions {
		if td.Dimension == nil {
			continue
		}

		valCount := len(td.Dimension.Values)
		if valCount == 0 {
			valCount = 1
		}

		expectedPerYear *= valCount
	}

	// Ambil fakta dari database
	filledCounts, err := s.factRepo.CountFactsByYear(table.ID, fromYear, toYear)
	if err != nil {
		return nil, fmt.Errorf("failed to count facts: %w", err)
	}

	// Ambil jumlah outlier per tahun
	outlierCounts, err := s.factRepo.CountOutliersByYear(table.ID, fromYear, toYear)
	if err != nil {
		return nil, fmt.Errorf("failed to count outliers: %w", err)
	}

	// Ambil jumlah revision per tahun
	revisionCounts, err := s.factRepo.CountRevisionsByYear(table.ID, fromYear, toYear)
	if err != nil {
		return nil, fmt.Errorf("failed to count revisions: %w", err)
	}

	// Persiapkan slice data
	data := make([]dto.DataInsightFact, 0, toYear-fromYear+1)

	totalExpecteds := 0
	totalFilleds := 0
	totalMissings := 0
	totalOutliers := 0
	totalRevisions := 0

	for year := fromYear; year <= toYear; year++ {
		filled := filledCounts[year]
		missing := expectedPerYear - filled
		if missing < 0 {
			missing = 0
		}

		outlier := outlierCounts[year]
		revision := revisionCounts[year]

		data = append(data, dto.DataInsightFact{
			Year:     year,
			Expected: expectedPerYear,
			Filled:   filled,
			Missing:  missing,
			Outlier:  outlier,
			Revision: revision,
		})

		totalExpecteds += expectedPerYear
		totalFilleds += filled
		totalMissings += missing
		totalOutliers += outlier
		totalRevisions += revision
	}

	return &dto.InsightFactsResponse{
		TableID:  table.ID,
		FromYear: fromYear,
		ToYear:   toYear,
		Summary: dto.SummaryInsightFacts{
			ExpectedPerYear: expectedPerYear,
			TotalExpecteds:  totalExpecteds,
			TotalFilleds:    totalFilleds,
			TotalMissings:   totalMissings,
			TotalOutliers:   totalOutliers,
			TotalRevisions:  totalRevisions,
		},
		Data: data,
	}, nil
}

func (s *FactService) GetInsightFactsForTables(
	tables []*models.Table,
	fromYear, toYear int,
) (map[string]*dto.InsightFactsResponse, error) {

	tableIDs := make([]string, len(tables))
	for i, t := range tables {
		tableIDs[i] = t.ID
	}

	// =============== QUERY SEKALI UNTUK SEMUA ===============
	filledCounts, err := s.factRepo.CountFactsByYearForTables(tableIDs, fromYear, toYear)
	if err != nil {
		return nil, err
	}

	outlierCounts, err := s.factRepo.CountOutliersForTables(tableIDs)
	if err != nil {
		return nil, err
	}

	revisionCounts, err := s.factRepo.CountRevisionsForTables(tableIDs)
	if err != nil {
		return nil, err
	}
	// ========================================================

	result := make(map[string]*dto.InsightFactsResponse, len(tables))

	for _, table := range tables {

		// Hitung expected per year
		expectedPerYear := 1
		for _, td := range table.Dimensions {
			if td.Dimension != nil {
				expectedPerYear *= len(td.Dimension.Values)
			}
		}

		data := make([]dto.DataInsightFact, 0, toYear-fromYear+1)
		totalExpecteds, totalFilleds, totalMissings := 0, 0, 0

		yearlyFilled := filledCounts[table.ID]

		for year := fromYear; year <= toYear; year++ {
			filled := 0
			if yearlyFilled != nil {
				filled = yearlyFilled[year]
			}

			missing := max(expectedPerYear-filled, 0)

			data = append(data, dto.DataInsightFact{
				Year:     year,
				Expected: expectedPerYear,
				Filled:   filled,
				Missing:  missing,
			})

			totalExpecteds += expectedPerYear
			totalFilleds += filled
			totalMissings += missing
		}

		// Ambil outlier, default 0
		totalOutliers := 0
		if v, ok := outlierCounts[table.ID]; ok {
			totalOutliers = v
		}

		// Ambil revision, default 0
		totalRevisions := 0
		if v, ok := revisionCounts[table.ID]; ok {
			totalRevisions = v
		}

		result[table.ID] = &dto.InsightFactsResponse{
			TableID:  table.ID,
			FromYear: fromYear,
			ToYear:   toYear,
			Summary: dto.SummaryInsightFacts{
				ExpectedPerYear: expectedPerYear,
				TotalExpecteds:  totalExpecteds,
				TotalFilleds:    totalFilleds,
				TotalMissings:   totalMissings,
				TotalOutliers:   totalOutliers,
				TotalRevisions:  totalRevisions,
			},
			Data: data,
		}
	}

	return result, nil
}

func (s *FactService) AnalyzeFacts(tableID string) error {
	facts, err := s.factRepo.FindAllByTable(tableID)
	if err != nil {
		return err
	}

	// 1. Kelompokkan facts berdasarkan DimensionValue yang sama
	groups := make(map[string][]*models.Fact) // key = dimensionKey
	for i := range facts {
		key := utils.BuildDimensionKey(facts[i])
		groups[key] = append(groups[key], facts[i])
	}

	// 2. Proses setiap group
	for _, groupFacts := range groups {

		// Ambil nilai Value saja (ignore nil)
		var values []float64
		var factIndex []int // index ke fact sebenarnya

		for idx, f := range groupFacts {
			if f.Value != nil {
				values = append(values, *f.Value)
				factIndex = append(factIndex, idx)
			}
		}

		// Tidak bisa analisa kalau <3 data
		if len(values) < 3 {
			continue
		}

		// 3. Jalankan Modified Z-Score
		outlierIndexes, mzScores, err := utils.DetectOutliersModifiedZ(values)
		if err != nil {
			return err
		}

		log.Printf("Modified Z-Scores for group (dimension key=%s): %v\n", utils.BuildDimensionKey(groupFacts[0]), mzScores)
		log.Printf("Found %d outliers in group (dimension key=%s) with %d values\n", len(outlierIndexes), utils.BuildDimensionKey(groupFacts[0]), len(values))

		// 4. Simpan hasil ke struct Fact
		outlierSet := make(map[int]bool)
		for _, oi := range outlierIndexes {
			outlierSet[oi] = true
		}

		for i, idx := range factIndex {
			isOutlier := outlierSet[i]
			groupFacts[idx].IsOutlier = &isOutlier
		}

	}

	// 5. Simpan semua fact yang sudah diperbarui
	for _, f := range facts {
		if err := s.factRepo.UpdateFact(f); err != nil {
			return err
		}
	}

	return nil
}

func (s *FactService) UnanalyzeFacts(tableID string) error {
	return s.factRepo.ResetOutliersByTable(tableID)
}

func (s *FactService) GetFactsByTableID(tableID string, dimValueIDs []string) ([]*dto.FactResponse, error) {
	facts, err := s.factRepo.FindAllByTableAndDimensionValues(tableID, dimValueIDs)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.FactResponse, 0, len(facts))
	for _, fact := range facts {
		responses = append(responses, mappers.ToFactResponse(fact))
	}

	return responses, nil
}
