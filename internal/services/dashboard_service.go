package services

import "statio/internal/dto"

type DashboardService struct {
	tableService *TableService
	factService  *FactService
}

func NewDashboardService(
	tableService *TableService,
	factService *FactService,
) *DashboardService {
	return &DashboardService{
		tableService: tableService,
		factService:  factService,
	}
}

func (s *DashboardService) GetDashboardStatistics(organizationId *string) (*dto.DashboardStatisticsResponse, error) {

	tables, err := s.tableService.GetTablesBase(&dto.FilterTablesRequest{
		OrganizationID: organizationId,
	})
	if err != nil {
		return nil, err
	}

	totalTables := len(tables)

	totalDraft := 0
	totalSubmitted := 0
	totalFinalized := 0

	for _, table := range tables {
		switch table.Status {
		case "draft":
			totalDraft++
		case "submitted":
			totalSubmitted++
		case "finalized":
			totalFinalized++
		}
	}

	response := &dto.DashboardStatisticsResponse{
		TotalTables:         totalTables,
		TotalTableDraft:     totalDraft,
		TotalTableSubmitted: totalSubmitted,
		TotalTableFinalized: totalFinalized,
	}

	return response, nil
}
