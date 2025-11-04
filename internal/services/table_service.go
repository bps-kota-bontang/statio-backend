package services

import (
	"fmt"
	"statio/internal/dto"
	"statio/internal/mappers"
	"statio/internal/repositories"

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

	responses := make([]*dto.TableListResponse, 0, len(tables))
	for _, dimension := range tables {
		responses = append(responses, mappers.ToTableListResponse(dimension))
	}

	return responses, total, nil
}

func (s *TableService) GetByID(id string, year int) (*dto.TableResponse, error) {
	table, err := s.tableRepo.FindByIDAndYear(id, year)
	if err != nil {
		return nil, err
	}

	response := mappers.ToTableResponse(table, year)
	return response, nil
}

func (s *TableService) UpdateTableFacts(tableID string, payload *dto.UpdateFactRequest) error {
	table, err := s.tableRepo.FindByIDForFactUpdate(tableID)
	if err != nil || table == nil {
		return fmt.Errorf("table not found")
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
