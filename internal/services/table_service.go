package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"statio/internal/dto"
	"statio/internal/mappers"
	"statio/internal/models"
	"statio/internal/repositories"
	"statio/utils"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type TableService struct {
	tableRepo    repositories.TableRepository
	factSvc      *FactService
	dimensionSvc *DimensionService
	asynqClient  *asynq.Client
	db           *gorm.DB
}

func NewTableService(
	tableRepo repositories.TableRepository,
	factSvc *FactService,
	dimensionSvc *DimensionService,
	asynqClient *asynq.Client,
	db *gorm.DB,
) *TableService {
	return &TableService{
		tableRepo:    tableRepo,
		factSvc:      factSvc,
		dimensionSvc: dimensionSvc,
		asynqClient:  asynqClient,
		db:           db,
	}
}

func (s *TableService) GetTablesBase(
	filter *dto.FilterTablesRequest,
) ([]*models.Table, error) {
	tables, err := s.tableRepo.FindTablesBase(filter)

	return tables, err
}

func (s *TableService) GetAllPaginated(
	search string,
	page, perPage int,
	sortBy, sortOrder string,
	filters map[string][]string,
) ([]*dto.TableListResponse, int64, error) {
	// 1) Ambil kandidat ringan
	lightTables, err := s.tableRepo.FindLight(search, sortBy, sortOrder, filters)
	if err != nil {
		return nil, 0, err
	}
	if len(lightTables) == 0 {
		return []*dto.TableListResponse{}, 0, nil
	}

	// Kumpulkan semua ID
	allIDs := make([]string, 0, len(lightTables))
	for _, t := range lightTables {
		allIDs = append(allIDs, t.ID)
	}

	// 2) Load dimensions untuk semua table candidates
	dimTables, err := s.tableRepo.LoadDimensionsForTableIDs(allIDs)
	if err != nil {
		return nil, 0, err
	}

	// Buat map ID -> table
	dimMap := make(map[string]*models.Table, len(dimTables))
	for _, t := range dimTables {
		dimMap[t.ID] = t
	}

	// Build list sesuai urutan allIDs
	tablesForCount := make([]*models.Table, 0, len(allIDs))
	for _, id := range allIDs {
		if t, ok := dimMap[id]; ok {
			tablesForCount = append(tablesForCount, t)
		} else {
			tablesForCount = append(tablesForCount, &models.Table{ID: id})
		}
	}

	// 3) Hitung missing facts untuk semua kandidat
	currentYear := time.Now().Year()
	fromYear := currentYear - 4
	toYear := currentYear - 1

	insightFactsMap, err := s.factSvc.GetInsightFactsForTables(tablesForCount, fromYear, toYear)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get missing facts: %w", err)
	}

	// 4) Filter missing_facts, outlier_facts, revision_facts (in-memory)

	// ================= Missing Facts Filter =================
	var filterHasMissing, filterNoMissing bool
	if vals, ok := filters["missing_facts"]; ok {
		for _, v := range vals {
			if v == "true" {
				filterHasMissing = true
			}
			if v == "false" {
				filterNoMissing = true
			}
		}
	}
	if filterHasMissing && filterNoMissing {
		filterHasMissing = false
		filterNoMissing = false
	}

	// ================= Outlier Facts Filter =================
	var filterHasOutlier, filterNoOutlier bool
	if vals, ok := filters["outlier_facts"]; ok {
		for _, v := range vals {
			if v == "true" {
				filterHasOutlier = true
			}
			if v == "false" {
				filterNoOutlier = true
			}
		}
	}
	if filterHasOutlier && filterNoOutlier {
		filterHasOutlier = false
		filterNoOutlier = false
	}

	// ================= Revision Facts Filter =================
	var filterHasRevision, filterNoRevision bool
	if vals, ok := filters["revision_facts"]; ok {
		for _, v := range vals {
			if v == "true" {
				filterHasRevision = true
			}
			if v == "false" {
				filterNoRevision = true
			}
		}
	}
	if filterHasRevision && filterNoRevision {
		filterHasRevision = false
		filterNoRevision = false
	}

	// ================= Final Filter Loop =================
	filteredIDs := make([]string, 0, len(allIDs))

	for _, id := range allIDs {

		missingTotal := 0
		outlierTotal := 0
		revisionTotal := 0
		if mf, ok := insightFactsMap[id]; ok {
			missingTotal = mf.Summary.TotalMissings
			outlierTotal = mf.Summary.TotalOutliers
			revisionTotal = mf.Summary.TotalRevisions
		}

		// Filtering rules
		if filterHasMissing && missingTotal == 0 {
			continue
		}
		if filterNoMissing && missingTotal > 0 {
			continue
		}

		if filterHasOutlier && outlierTotal == 0 {
			continue
		}
		if filterNoOutlier && outlierTotal > 0 {
			continue
		}

		if filterHasRevision && revisionTotal == 0 {
			continue
		}
		if filterNoRevision && revisionTotal > 0 {
			continue
		}

		filteredIDs = append(filteredIDs, id)
	}

	// ============== 5) Total setelah filter ==============
	total := int64(len(filteredIDs))

	// ============== 5.5) Sorting Virtual Fields ==============
	if sortBy == "missing_facts" || sortBy == "outlier_facts" || sortBy == "revision_facts" {
		sort.Slice(filteredIDs, func(i, j int) bool {
			a := insightFactsMap[filteredIDs[i]]
			b := insightFactsMap[filteredIDs[j]]

			var va, vb int

			switch sortBy {
			case "missing_facts":
				if a != nil {
					va = a.Summary.TotalMissings
				}
				if b != nil {
					vb = b.Summary.TotalMissings
				}
			case "outlier_facts":
				if a != nil {
					va = a.Summary.TotalOutliers
				}
				if b != nil {
					vb = b.Summary.TotalOutliers
				}
			case "revision_facts":
				if a != nil {
					va = a.Summary.TotalRevisions
				}
				if b != nil {
					vb = b.Summary.TotalRevisions
				}
			}

			if sortOrder == "desc" {
				return va > vb
			}
			return va < vb
		})
	}

	// ============== 6) Pagination in-memory ==============
	start := (page - 1) * perPage
	if start >= len(filteredIDs) {
		return []*dto.TableListResponse{}, total, nil
	}
	end := min(start+perPage, len(filteredIDs))
	pagedIDs := filteredIDs[start:end]

	// 7) Load detail hanya untuk ID yang tampil
	detailedTables, err := s.tableRepo.FindByIDsDetailed(pagedIDs)
	if err != nil {
		return nil, 0, err
	}

	detailMap := make(map[string]*models.Table, len(detailedTables))
	for _, t := range detailedTables {
		detailMap[t.ID] = t
	}

	// 8) Build response
	responses := make([]*dto.TableListResponse, 0, len(pagedIDs))
	for _, id := range pagedIDs {
		if t := detailMap[id]; t != nil {
			resp := mappers.ToTableListResponse(t)
			if mf, ok := insightFactsMap[id]; ok {
				resp.InsightFactsSummary = &mf.Summary
			}
			responses = append(responses, resp)
		}
	}

	return responses, total, nil
}

func (s *TableService) GetByID(id string, year *int) (*dto.TableResponse, error) {
	var (
		table *models.Table
		err   error
	)

	// Default: gunakan year yang diberikan
	useYear := year

	countDimensions, err := s.tableRepo.CountDimensionsByTableID(id)
	if err != nil {
		return nil, err
	}

	if year != nil {
		// Jika tidak ada dimension (nil atau 0), abaikan year
		if countDimensions == nil || *countDimensions == 0 {
			useYear = nil
		}
	} else {
		// Jika year tidak diberikan, tapi table punya dimension, set useYear ke last year
		if countDimensions != nil && *countDimensions > 0 {
			lastYear := time.Now().Year() - 1
			useYear = &lastYear
		}
	}

	// Ambil tabel sesuai kondisi
	if useYear == nil {
		table, err = s.tableRepo.FindDetailedByID(id)
	} else {
		table, err = s.tableRepo.FindByIDAndYear(id, *useYear)
	}
	if err != nil {
		return nil, err
	}

	response := mappers.ToTableResponse(table, useYear)
	return response, nil
}

func (s *TableService) UpdateTableFacts(tableID string, payload *dto.UpdateFactRequest, roles []string, organizationID *string) error {
	table, err := s.tableRepo.FindForFactUpdate(tableID)
	if err != nil || table == nil {
		return fmt.Errorf("table not found")
	}

	if !utils.IsAdmin(roles) {
		if organizationID == nil || table.OrganizationID == nil || *organizationID != *table.OrganizationID {
			return fmt.Errorf("you are not authorized to update facts for this table")
		}
	}

	return s.factSvc.SaveOrUpdateFacts(table, payload)
}

func (s *TableService) GetTableFacts(tableID string, dimValueIDs, roles []string, organizationID *string) ([]*dto.FactResponse, error) {
	table, err := s.tableRepo.FindForFactUpdate(tableID)
	if err != nil || table == nil {
		return nil, fmt.Errorf("table not found")
	}

	if !utils.IsAdmin(roles) {
		if organizationID == nil || table.OrganizationID == nil || *organizationID != *table.OrganizationID {
			return nil, fmt.Errorf("you are not authorized to view facts for this table")
		}
	}

	return s.factSvc.GetFactsByTableID(tableID, dimValueIDs)
}

func (s *TableService) Create(input *dto.CreateTableRequest) (*dto.TableListResponse, error) {
	var result *dto.TableListResponse

	// Mulai transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Validasi semua DimensionID ada
	if err := s.dimensionSvc.ValidateIDs(input.DimensionIDs); err != nil {
		return nil, err
	}

	// 2. Buat Table baru
	table := mappers.ToTableModel(input)

	// 3. Simpan Table + TableDimension lewat repository
	if err := s.tableRepo.CreateWithTx(tx, table); err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	result = mappers.ToTableListResponse(table)
	return result, nil
}

func (s *TableService) UpdateWithRelations(id string, input *dto.UpdateTableRequest) error {
	table, err := s.tableRepo.FindBaseByID(id)
	if err != nil {
		return err
	}

	mappers.ApplyTableUpdateFromRequest(table, input)

	return s.tableRepo.UpdateWithRelations(table, input.DimensionIDs)
}

func (s *TableService) AssignOrganizationBulk(organizationID string, tableIDs []string) error {
	return s.tableRepo.UpdateOrganizationBulk(organizationID, tableIDs)
}

func (s *TableService) AddLabelsBulk(
	input *dto.AddLabelsToTablesRequest,
	roles []string, organizationID *string,
) error {
	if !utils.IsAdmin(roles) {
		// Cek setiap table apakah boleh diakses
		tables, err := s.tableRepo.FindByIDs(input.TableIDs)
		if err != nil {
			return err
		}

		for _, table := range tables {
			if organizationID == nil || table.OrganizationID == nil || *organizationID != *table.OrganizationID {
				return fmt.Errorf("you are not authorized to add labels to one or more of the specified tables")
			}
		}
	}

	return s.tableRepo.AddLabelsBulk(input.Labels, input.TableIDs)
}

func (s *TableService) GetAllLabels() ([]*dto.TableLabelResponse, error) {
	labels, err := s.tableRepo.FindAllLabels()
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.TableLabelResponse, 0, len(labels))
	for _, label := range labels {
		responses = append(responses, mappers.ToTableLabelResponse(*label))
	}

	return responses, nil
}

func (s *TableService) UpdateTableLabels(
	tableID string,
	input *dto.UpdateTableLabelsRequest,
	roles []string, organizationID *string,
) error {
	if !utils.IsAdmin(roles) {
		table, err := s.tableRepo.FindBaseByID(tableID)
		if err != nil {
			return err
		}

		if organizationID == nil || table.OrganizationID == nil || *organizationID != *table.OrganizationID {
			return fmt.Errorf("you are not authorized to update labels for this table")
		}
	}

	return s.tableRepo.UpdateLabels(tableID, input.Labels)
}

func (s *TableService) UpdateTableName(
	tableID string,
	newName string,
	roles []string, organizationID *string,
) error {
	table, err := s.tableRepo.FindBaseByID(tableID)
	if err != nil {
		return err
	}

	if !utils.IsAdmin(roles) {
		if organizationID == nil || table.OrganizationID == nil || *organizationID != *table.OrganizationID {
			return fmt.Errorf("you are not authorized to update the name for this table")
		}
	}

	table.Name = newName

	return s.tableRepo.Update(table)
}

func (s *TableService) UpdateTableNotes(
	tableID string,
	newNotes *string,
	roles []string, organizationID *string,
) error {
	table, err := s.tableRepo.FindBaseByID(tableID)
	if err != nil {
		return err
	}

	if !utils.IsAdmin(roles) {
		if organizationID == nil || table.OrganizationID == nil || *organizationID != *table.OrganizationID {
			return fmt.Errorf("you are not authorized to update the notes for this table")
		}
	}

	table.Notes = newNotes

	return s.tableRepo.Update(table)
}

func (s *TableService) UpdateTableIsLocked(
	tableID string,
	isLocked bool,
) error {
	table, err := s.tableRepo.FindBaseByID(tableID)
	if err != nil {
		return err
	}

	table.IsLocked = isLocked

	return s.tableRepo.Update(table)
}

func (s *TableService) UpdateTableStatus(
	tableID string,
	status string,
	roles []string,
	organizationID *string,
) error {
	table, err := s.tableRepo.FindBaseByID(tableID)
	if err != nil {
		return err
	}

	if !utils.IsAdmin(roles) {
		if organizationID == nil || table.OrganizationID == nil || *organizationID != *table.OrganizationID {
			return fmt.Errorf("you are not authorized to update the status for this table")
		}
	}

	if status == "draft" {
		if !utils.IsAdmin(roles) {
			return fmt.Errorf("only admin users can change the table status to draft")
		}
		table.IsLocked = false

		payload, err := json.Marshal(&dto.UnanalyzeFactPayload{
			TableID: table.ID,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal unanalyze fact payload: %w", err)
		}

		task := asynq.NewTask("fact:unanalyze", payload)

		if _, err := s.asynqClient.Enqueue(task); err != nil {
			return fmt.Errorf("failed to enqueue unanalyze fact task: %w", err)
		}
	}

	if status == "submitted" {
		table.IsLocked = true

		payload, err := json.Marshal(&dto.AnalyzeFactPayload{
			TableID: table.ID,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal analyze fact payload: %w", err)
		}

		task := asynq.NewTask("fact:analyze", payload)

		if _, err := s.asynqClient.Enqueue(task); err != nil {
			return fmt.Errorf("failed to enqueue analyze fact task: %w", err)
		}
	}

	if status == "finalized" {
		if !utils.IsAdmin(roles) {
			return fmt.Errorf("only admin users can change the table status to finalized")
		}
		table.IsLocked = true

		payload, err := json.Marshal(&dto.AnalyzeFactPayload{
			TableID: table.ID,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal analyze fact payload: %w", err)
		}

		task := asynq.NewTask("fact:analyze", payload)

		if _, err := s.asynqClient.Enqueue(task); err != nil {
			return fmt.Errorf("failed to enqueue analyze fact task: %w", err)
		}
	}

	table.Status = status

	return s.tableRepo.Update(table)
}

func (s *TableService) GetInsightFacts(
	tableID string,
	roles []string,
	organizationID *string,
	fromYear, toYear int,
) (*dto.InsightFactsResponse, error) {
	table, err := s.tableRepo.FindDetailedByID(tableID)
	if err != nil || table == nil {
		return nil, fmt.Errorf("table not found")
	}

	if !utils.IsAdmin(roles) {
		if organizationID == nil || table.OrganizationID == nil || *organizationID != *table.OrganizationID {
			return nil, fmt.Errorf("you are not authorized to view missing facts for this table")
		}
	}

	responses, err := s.factSvc.GetInsightFactsForTable(table, fromYear, toYear)
	if err != nil {
		return nil, err
	}

	return responses, nil
}

func (s *TableService) AnalyzeTable(tableID string) error {
	payload, err := json.Marshal(&dto.AnalyzeFactPayload{
		TableID: tableID,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal analyze fact payload: %w", err)
	}

	task := asynq.NewTask("fact:analyze", payload)

	if _, err := s.asynqClient.Enqueue(task); err != nil {
		return fmt.Errorf("failed to enqueue analyze fact task: %w", err)
	}

	return nil
}

func (s *TableService) AnalyzeTables(tableIDs []string) error {
	for _, tableID := range tableIDs {
		if err := s.AnalyzeTable(tableID); err != nil {
			log.Printf("failed to enqueue analyze task for table %s: %v", tableID, err)
		}
	}

	return nil
}

func (s *TableService) CommitTable(tableID string) error {
	payload, err := json.Marshal(&dto.CommitFactPayload{
		TableID: tableID,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal analyze fact payload: %w", err)
	}

	task := asynq.NewTask("fact:commit", payload)

	if _, err := s.asynqClient.Enqueue(task); err != nil {
		return fmt.Errorf("failed to enqueue commit fact task: %w", err)
	}

	return nil
}

func (s *TableService) CommitTables(tableIDs []string) error {
	for _, tableID := range tableIDs {
		if err := s.CommitTable(tableID); err != nil {
			log.Printf("failed to commit facts for table %s: %v", tableID, err)
		}
	}

	return nil
}

func (s *TableService) ExportTable(tableID string, years []int) (*dto.TableExportResponse, error) {
	tableModel, err := s.tableRepo.FindByIDAndMultiYear(tableID, years)
	if err != nil {
		return nil, err
	}
	table := mappers.ToTableResponse(tableModel, nil)

	// Create a new Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("failed to close excel file: %v", err)
		}
	}()

	// Sort years
	sort.Ints(years)

	// Create header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#D3D3D3"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	// Write data rows
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	// Check dimension count for different formats
	if len(table.Dimensions) == 0 {
		// Format for 0 dimension - organization and multiple year columns
		sheetName := table.Name
		f.SetSheetName("Sheet1", sheetName)
		currentRow := 1

		// Write "Kabupaten/Kota" header in A1:A2 merged
		headerRow := currentRow
		cellA1, _ := excelize.CoordinatesToCellName(1, headerRow)
		cellA2, _ := excelize.CoordinatesToCellName(1, headerRow+1)
		f.SetCellValue(sheetName, cellA1, "Kabupaten/Kota")
		f.MergeCell(sheetName, cellA1, cellA2)
		f.SetCellStyle(sheetName, cellA1, cellA2, headerStyle)

		// Write "Tahun" header merged across all year columns (B1 to lastCol1)
		cellB1, _ := excelize.CoordinatesToCellName(2, headerRow)
		cellLast1, _ := excelize.CoordinatesToCellName(len(years)+1, headerRow)
		f.SetCellValue(sheetName, cellB1, "Tahun")
		f.MergeCell(sheetName, cellB1, cellLast1)
		f.SetCellStyle(sheetName, cellB1, cellLast1, headerStyle)
		currentRow++

		// Write year values in B2, C2, D2, ...
		for colIdx, year := range years {
			cellName, _ := excelize.CoordinatesToCellName(colIdx+2, currentRow)
			f.SetCellValue(sheetName, cellName, year)
			f.SetCellStyle(sheetName, cellName, cellName, headerStyle)
		}
		currentRow++

		// Create map: year -> value
		yearValues := make(map[int]interface{})
		for _, fact := range table.Facts {
			if fact.Value != nil {
				yearValues[fact.Year] = *fact.Value
			}
		}

		// Write organization name in A3
		cellA3, _ := excelize.CoordinatesToCellName(1, currentRow)
		orgName := "Kota Bontang"
		f.SetCellValue(sheetName, cellA3, orgName)
		f.SetCellStyle(sheetName, cellA3, cellA3, dataStyle)

		// Write values in B3, C3, D3, ...
		for colIdx, year := range years {
			cellName, _ := excelize.CoordinatesToCellName(colIdx+2, currentRow)
			if yearValues[year] != nil {
				f.SetCellValue(sheetName, cellName, yearValues[year])
			} else {
				f.SetCellValue(sheetName, cellName, "")
			}
			f.SetCellStyle(sheetName, cellName, cellName, dataStyle)
		}

		// Auto-fit columns
		f.SetColWidth(sheetName, "A", "A", 25)
		for i := range years {
			colName, _ := excelize.ColumnNumberToName(i + 2)
			f.SetColWidth(sheetName, colName, colName, 15)
		}
	} else if len(table.Dimensions) == 1 {
		// Format for 1 dimension - horizontal layout with years as columns
		sheetName := table.Name
		f.SetSheetName("Sheet1", sheetName)
		currentRow := 1

		dim := table.Dimensions[0]

		// Collect dimension values (preserve order from dimension definition)
		dimValues := []string{}
		for _, dimValue := range dim.Values {
			dimValues = append(dimValues, dimValue.Name)
		}

		// Write dimension name header (merged cells A1:A2)
		headerRow := currentRow
		cellA1, _ := excelize.CoordinatesToCellName(1, headerRow)
		cellA2, _ := excelize.CoordinatesToCellName(1, headerRow+1)
		f.SetCellValue(sheetName, cellA1, dim.Name)
		f.MergeCell(sheetName, cellA1, cellA2)
		f.SetCellStyle(sheetName, cellA1, cellA2, headerStyle)

		// Write "Tahun" header merged across all year columns (B1 to lastCol1)
		cellB1, _ := excelize.CoordinatesToCellName(2, headerRow)
		cellLast1, _ := excelize.CoordinatesToCellName(len(years)+1, headerRow)
		f.SetCellValue(sheetName, cellB1, "Tahun")
		f.MergeCell(sheetName, cellB1, cellLast1)
		f.SetCellStyle(sheetName, cellB1, cellLast1, headerStyle)
		currentRow++

		// Write year values in row 2 (B2, C2, D2, ...)
		for colIdx, year := range years {
			cellName, _ := excelize.CoordinatesToCellName(colIdx+2, currentRow)
			f.SetCellValue(sheetName, cellName, year)
			f.SetCellStyle(sheetName, cellName, cellName, headerStyle)
		}
		currentRow++

		// Create pivot data structure: dimValue -> year -> value
		pivotData := make(map[string]map[int]interface{})
		for _, fact := range table.Facts {
			var dimVal string
			for _, fv := range fact.Dimensions {
				if fv.ID == dim.ID {
					dimVal = fv.Value.Name
					break
				}
			}
			if pivotData[dimVal] == nil {
				pivotData[dimVal] = make(map[int]interface{})
			}
			if fact.Value != nil {
				pivotData[dimVal][fact.Year] = *fact.Value
			}
		}

		// Write data rows (A3 onwards: dimension values, B3 onwards: values)
		for _, dimVal := range dimValues {
			// Row header (dimension value in column A)
			cellName, _ := excelize.CoordinatesToCellName(1, currentRow)
			f.SetCellValue(sheetName, cellName, dimVal)
			f.SetCellStyle(sheetName, cellName, cellName, dataStyle)

			// Data cells for each year (B, C, D, ...)
			for colIdx, year := range years {
				cellName, _ := excelize.CoordinatesToCellName(colIdx+2, currentRow)
				if pivotData[dimVal] != nil && pivotData[dimVal][year] != nil {
					f.SetCellValue(sheetName, cellName, pivotData[dimVal][year])
				} else {
					f.SetCellValue(sheetName, cellName, "")
				}
				f.SetCellStyle(sheetName, cellName, cellName, dataStyle)
			}
			currentRow++
		}

		// Auto-fit columns
		f.SetColWidth(sheetName, "A", "A", 25)
		for i := 0; i < len(years); i++ {
			colName, _ := excelize.ColumnNumberToName(i + 2)
			f.SetColWidth(sheetName, colName, colName, 15)
		}
	} else if len(table.Dimensions) == 2 {
		// Format for 2 dimensions - each year gets its own sheet
		dim1 := table.Dimensions[0] // Row dimension (vertical)
		dim2 := table.Dimensions[1] // Column dimension (horizontal)

		// Use dimension values from table definition (preserve order)
		dim1Values := []string{}
		dim2Values := []string{}

		for _, dimValue := range dim1.Values {
			dim1Values = append(dim1Values, dimValue.Name)
		}

		for _, dimValue := range dim2.Values {
			dim2Values = append(dim2Values, dimValue.Name)
		}

		// Create pivot data structure by year: year -> dim1Val -> dim2Val -> value
		pivotDataByYear := make(map[int]map[string]map[string]interface{})
		for _, fact := range table.Facts {
			var dim1Val, dim2Val string
			for _, fv := range fact.Dimensions {
				if fv.ID == dim1.ID {
					dim1Val = fv.Value.Name
				}
				if fv.ID == dim2.ID {
					dim2Val = fv.Value.Name
				}
			}
			if pivotDataByYear[fact.Year] == nil {
				pivotDataByYear[fact.Year] = make(map[string]map[string]interface{})
			}
			if pivotDataByYear[fact.Year][dim1Val] == nil {
				pivotDataByYear[fact.Year][dim1Val] = make(map[string]interface{})
			}
			if fact.Value != nil {
				pivotDataByYear[fact.Year][dim1Val][dim2Val] = *fact.Value
			}
		}

		// Delete default Sheet1 after creating all year sheets
		firstSheet := true

		// Create a sheet for each year
		for _, year := range years {
			sheetName := fmt.Sprintf("%d (Tahunan)", year)

			if firstSheet {
				// Rename Sheet1 for the first year
				f.SetSheetName("Sheet1", sheetName)
				firstSheet = false
			} else {
				// Create new sheet for other years
				_, err := f.NewSheet(sheetName)
				if err != nil {
					log.Printf("failed to create sheet for year %d: %v", year, err)
					continue
				}
			}

			currentRow := 1
			pivotData := pivotDataByYear[year]

			// Row 1: A1:A2 merged with dim1.Name (Dimension 1)
			cellA1, _ := excelize.CoordinatesToCellName(1, currentRow)
			cellA2, _ := excelize.CoordinatesToCellName(1, currentRow+1)
			f.SetCellValue(sheetName, cellA1, dim1.Name)
			f.MergeCell(sheetName, cellA1, cellA2)
			f.SetCellStyle(sheetName, cellA1, cellA2, headerStyle)

			// Row 1: B1 to [lastCol]1 merged with dim2.Name (Dimension 2)
			cellB1, _ := excelize.CoordinatesToCellName(2, currentRow)
			cellLast1, _ := excelize.CoordinatesToCellName(len(dim2Values)+1, currentRow)
			f.SetCellValue(sheetName, cellB1, dim2.Name)
			f.MergeCell(sheetName, cellB1, cellLast1)
			f.SetCellStyle(sheetName, cellB1, cellLast1, headerStyle)
			currentRow++

			// Row 2: Write dim2 values starting from B2
			for colIdx, dim2Val := range dim2Values {
				cellName, _ := excelize.CoordinatesToCellName(colIdx+2, currentRow)
				f.SetCellValue(sheetName, cellName, dim2Val)
				f.SetCellStyle(sheetName, cellName, cellName, headerStyle)
			}
			currentRow++

			// Row 3+: Write dim1 values in column A and data grid
			for _, dim1Val := range dim1Values {
				// Column A: dimension 1 value
				cellName, _ := excelize.CoordinatesToCellName(1, currentRow)
				f.SetCellValue(sheetName, cellName, dim1Val)
				f.SetCellStyle(sheetName, cellName, cellName, headerStyle)

				// Data cells (B, C, D, E, ...)
				for colIdx, dim2Val := range dim2Values {
					cellName, _ := excelize.CoordinatesToCellName(colIdx+2, currentRow)
					if pivotData != nil && pivotData[dim1Val] != nil && pivotData[dim1Val][dim2Val] != nil {
						f.SetCellValue(sheetName, cellName, pivotData[dim1Val][dim2Val])
					} else {
						f.SetCellValue(sheetName, cellName, "")
					}
					f.SetCellStyle(sheetName, cellName, cellName, dataStyle)
				}
				currentRow++
			}

			// Auto-fit columns
			f.SetColWidth(sheetName, "A", "A", 25)
			for i := 0; i < len(dim2Values); i++ {
				colName, _ := excelize.ColumnNumberToName(i + 2)
				f.SetColWidth(sheetName, colName, colName, 15)
			}
		}
	} else {
		// Standard table format for tables with 3+ dimensions
		sheetName := table.Name
		f.SetSheetName("Sheet1", sheetName)
		currentRow := 1

		// Build headers
		headers := []string{"No"}
		for _, dim := range table.Dimensions {
			headers = append(headers, dim.Name)
		}
		headers = append(headers, "Year", "Value")

		// Write headers
		headerRow := currentRow
		for colIdx, header := range headers {
			cellName, _ := excelize.CoordinatesToCellName(colIdx+1, headerRow)
			f.SetCellValue(sheetName, cellName, header)
			f.SetCellStyle(sheetName, cellName, cellName, headerStyle)
		}
		currentRow++

		// Standard table format
		for idx, fact := range table.Facts {
			colIdx := 0

			// Row number
			cellName, _ := excelize.CoordinatesToCellName(colIdx+1, currentRow)
			f.SetCellValue(sheetName, cellName, idx+1)
			f.SetCellStyle(sheetName, cellName, cellName, dataStyle)
			colIdx++

			// Dimension values
			for _, dim := range table.Dimensions {
				var dimValue string
				for _, fv := range fact.Dimensions {
					if fv.ID == dim.ID {
						dimValue = fv.Value.Name
						break
					}
				}
				cellName, _ := excelize.CoordinatesToCellName(colIdx+1, currentRow)
				f.SetCellValue(sheetName, cellName, dimValue)
				f.SetCellStyle(sheetName, cellName, cellName, dataStyle)
				colIdx++
			}

			// Year
			cellName, _ = excelize.CoordinatesToCellName(colIdx+1, currentRow)
			f.SetCellValue(sheetName, cellName, fact.Year)
			f.SetCellStyle(sheetName, cellName, cellName, dataStyle)
			colIdx++

			// Fact value
			cellName, _ = excelize.CoordinatesToCellName(colIdx+1, currentRow)
			if fact.Value != nil {
				f.SetCellValue(sheetName, cellName, *fact.Value)
			} else {
				f.SetCellValue(sheetName, cellName, "")
			}
			f.SetCellStyle(sheetName, cellName, cellName, dataStyle)

			currentRow++
		}

		// Auto-fit columns
		for i := 0; i < len(headers); i++ {
			colName, _ := excelize.ColumnNumberToName(i + 1)
			f.SetColWidth(sheetName, colName, colName, 15)
		}
	}

	// Save to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write excel file: %w", err)
	}

	file := buf.Bytes()

	// Generate filename with all years listed
	var filename string
	if len(years) == 1 {
		filename = fmt.Sprintf("%s_%d.xlsx", table.Name, years[0])
	} else {
		// List all years separated by underscores
		yearStrs := make([]string, len(years))
		for i, year := range years {
			yearStrs[i] = fmt.Sprintf("%d", year)
		}
		filename = fmt.Sprintf("%s_%s.xlsx", table.Name, strings.Join(yearStrs, "_"))
	}

	return &dto.TableExportResponse{
		Name: filename,
		File: file,
	}, nil
}
