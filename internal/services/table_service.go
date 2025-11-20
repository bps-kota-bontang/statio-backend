package services

import (
	"fmt"
	"statio/internal/dto"
	"statio/internal/mappers"
	"statio/internal/models"
	"statio/internal/repositories"
	"statio/utils"
	"time"

	"gorm.io/gorm"
)

type TableService struct {
	tableRepo    repositories.TableRepository
	factSvc      *FactService
	dimensionSvc *DimensionService
	db           *gorm.DB
}

func NewTableService(
	tableRepo repositories.TableRepository,
	factSvc *FactService,
	dimensionSvc *DimensionService,
	db *gorm.DB,
) *TableService {
	return &TableService{
		tableRepo:    tableRepo,
		factSvc:      factSvc,
		dimensionSvc: dimensionSvc,
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
	// 1) Ambil kandidat (light)
	lightTables, err := s.tableRepo.FindLight(search, sortBy, sortOrder, filters)
	if err != nil {
		return nil, 0, err
	}
	if len(lightTables) == 0 {
		return []*dto.TableListResponse{}, 0, nil
	}

	// collect all IDs
	allIDs := make([]string, 0, len(lightTables))
	for _, t := range lightTables {
		allIDs = append(allIDs, t.ID)
	}

	// 2) Load dimensions+values for all candidate IDs (dipakai hitung expectedPerYear)
	dimTables, err := s.tableRepo.LoadDimensionsForTableIDs(allIDs)
	if err != nil {
		return nil, 0, err
	}

	// buat map id -> *models.Table (with dimensions loaded)
	dimMap := make(map[string]*models.Table, len(dimTables))
	for _, t := range dimTables {
		dimMap[t.ID] = t
	}

	// gunakan dimMap to build list of tables for fact counting (reuse GetMissingFactsForTables)
	// Build a slice in same order as allIDs
	tablesForCount := make([]*models.Table, 0, len(allIDs))
	for _, id := range allIDs {
		if t, ok := dimMap[id]; ok {
			tablesForCount = append(tablesForCount, t)
		} else {
			// if not loaded for some reason, create a minimal table with ID
			tablesForCount = append(tablesForCount, &models.Table{ID: id})
		}
	}

	// 3) Hitung missing facts untuk semua candidate tables
	currentYear := time.Now().Year()
	fromYear := currentYear - 4
	toYear := currentYear - 1

	missingFactsMap, err := s.factSvc.GetMissingFactsForTables(tablesForCount, fromYear, toYear)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get missing facts: %w", err)
	}

	// 4) Apply missing_facts filter in memory (if requested)
	var missingFilter *bool
	if v, ok := filters["missing_facts"]; ok && len(v) > 0 {
		switch v[0] {
		case "true":
			x := true
			missingFilter = &x
		case "false":
			x := false
			missingFilter = &x
		}
	}

	// Build filtered ID list preserving original order
	filteredIDs := make([]string, 0, len(allIDs))
	for _, id := range allIDs {
		sum := 0
		if mf, ok := missingFactsMap[id]; ok {
			sum = mf.Summary.TotalMissing
		} else {
			// if not present assume missing? you can decide; here assume missing (safe)
			sum = 0
		}

		if missingFilter != nil {
			if *missingFilter && sum == 0 {
				continue
			}
			if !*missingFilter && sum > 0 {
				continue
			}
		}
		filteredIDs = append(filteredIDs, id)
	}

	// 5) Total after filters
	total := int64(len(filteredIDs))

	// 6) Pagination in-memory
	start := (page - 1) * perPage
	if start >= len(filteredIDs) {
		return []*dto.TableListResponse{}, total, nil
	}
	end := start + perPage
	if end > len(filteredIDs) {
		end = len(filteredIDs)
	}
	pagedIDs := filteredIDs[start:end]

	// 7) Load detailed info only for pagedIDs
	detailedTables, err := s.tableRepo.FindByIDsDetailed(pagedIDs)
	if err != nil {
		return nil, 0, err
	}

	// Make map for easy lookup
	detailMap := make(map[string]*models.Table, len(detailedTables))
	for _, t := range detailedTables {
		detailMap[t.ID] = t
	}

	// 8) Build response preserving pagedIDs order
	responses := make([]*dto.TableListResponse, 0, len(pagedIDs))
	for _, id := range pagedIDs {
		t := detailMap[id]
		if t == nil {
			// skip if somehow missing
			continue
		}
		resp := mappers.ToTableListResponse(t)
		if mf, ok := missingFactsMap[id]; ok {
			resp.MissingFactsSummary = &mf.Summary
		}
		responses = append(responses, resp)
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
