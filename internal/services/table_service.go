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

	var total int64
	if err := s.tableRepo.Count(search, filters, &total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	tables, err := s.tableRepo.FindPaginated(search, perPage, offset, sortBy, sortOrder, filters)
	if err != nil {
		return nil, 0, err
	}

	// -----------------------------
	// HITUNG MISSING FACTS SEMUA TABLE
	// -----------------------------
	currentYear := time.Now().Year()
	fromYear := currentYear - 4
	toYear := currentYear - 1

	missingFactsMap, err := s.factSvc.GetMissingFactsForTables(tables, fromYear, toYear)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get missing facts: %w", err)
	}

	// ------------------------------------------
	//  FILTER MISSING FACTS (dari filters map)
	// ------------------------------------------

	var missingFactsFilter *bool // nil = no filter
	if values, ok := filters["missing_facts"]; ok && len(values) > 0 {
		// gunakan hanya index 0
		switch values[0] {
		case "true":
			tmp := true
			missingFactsFilter = &tmp
		case "false":
			tmp := false
			missingFactsFilter = &tmp
		}
	}

	filteredTables := make([]*models.Table, 0)
	for _, table := range tables {
		summary := missingFactsMap[table.ID].Summary
		totalMissing := summary.TotalMissing

		// Jika ada filter missing_facts, apply filter
		if missingFactsFilter != nil {
			if *missingFactsFilter && totalMissing == 0 {
				continue
			}
			if !*missingFactsFilter && totalMissing > 0 {
				continue
			}
		}

		filteredTables = append(filteredTables, table)
	}

	// total baru setelah filter
	total = int64(len(filteredTables))

	// -----------------------------
	// BANGUN RESPONSE
	// -----------------------------
	responses := make([]*dto.TableListResponse, 0, len(filteredTables))

	for _, table := range filteredTables {
		resp := mappers.ToTableListResponse(table)

		if missing, ok := missingFactsMap[table.ID]; ok {
			resp.MissingFactsSummary = &missing.Summary
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
