package services

import (
	"encoding/json"
	"fmt"
	"log"
	"statio/internal/dto"
	"statio/internal/mappers"
	"statio/internal/models"
	"statio/internal/repositories"
	"statio/utils"
	"time"

	"github.com/hibiken/asynq"
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

func (s *TableService) GetAll() ([]*dto.TableListResponse, error) {
	tables, err := s.tableRepo.FindAll()
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.TableListResponse, 0, len(tables))
	for _, table := range tables {
		responses = append(responses, mappers.ToTableListResponse(table))
	}

	return responses, nil
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

	missingFactsMap, err := s.factSvc.GetMissingFactsForTables(tablesForCount, fromYear, toYear)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get missing facts: %w", err)
	}

	outliersMap, err := s.factSvc.GetOutlierCounts(allIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count outliers: %w", err)
	}

	revisionMap, err := s.factSvc.GetRevisionCounts(allIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count revisions: %w", err)
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

		// missing facts
		missingTotal := 0
		if mf, ok := missingFactsMap[id]; ok {
			missingTotal = mf.Summary.TotalMissing
		}

		// outliers
		outlierTotal := 0
		if o, ok := outliersMap[id]; ok {
			outlierTotal = o.TotalOutliers
		}

		// revisions
		revisionTotal := 0
		if rv, ok := revisionMap[id]; ok {
			revisionTotal = rv.TotalRevisions
		}

		// ==== Apply filters ====

		// missing_facts filter
		if filterHasMissing && missingTotal == 0 {
			continue
		}
		if filterNoMissing && missingTotal > 0 {
			continue
		}

		// outlier_facts filter
		if filterHasOutlier && outlierTotal == 0 {
			continue
		}
		if filterNoOutlier && outlierTotal > 0 {
			continue
		}

		// revision_facts filter
		if filterHasRevision && revisionTotal == 0 {
			continue
		}
		if filterNoRevision && revisionTotal > 0 {
			continue
		}

		// lolos semua filter
		filteredIDs = append(filteredIDs, id)
	}

	// 5) Hitung total setelah filter
	total := int64(len(filteredIDs))

	// 6) Pagination in-memory
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
			if mf, ok := missingFactsMap[id]; ok {
				resp.MissingFactsSummary = &mf.Summary
			}
			if outlierCount, ok := outliersMap[id]; ok {
				resp.OutlierFactsSummary = outlierCount
			}
			if revisionCount, ok := revisionMap[id]; ok {
				resp.RevisionFactsSummary = revisionCount
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
	}

	table.Status = status

	return s.tableRepo.Update(table)
}

func (s *TableService) GetMissingFacts(
	tableID string,
	roles []string,
	organizationID *string,
	fromYear, toYear int,
) (*dto.MissingFactsResponse, error) {
	table, err := s.tableRepo.FindDetailedByID(tableID)
	if err != nil || table == nil {
		return nil, fmt.Errorf("table not found")
	}

	if !utils.IsAdmin(roles) {
		if organizationID == nil || table.OrganizationID == nil || *organizationID != *table.OrganizationID {
			return nil, fmt.Errorf("you are not authorized to view missing facts for this table")
		}
	}

	responses, err := s.factSvc.GetMissingFactsForTable(table, fromYear, toYear)
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
