package services

import (
	"fmt"
	"sort"
	"statio/internal/dto"
	"time"
)

var (
	// Collection period dates
	CollectionStartDate = time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	CollectionEndDate   = time.Date(2026, 2, 13, 0, 0, 0, 0, time.UTC)
)

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

func (s *DashboardService) GetOrganizationCompletion() ([]dto.OrganizationCompletionResponse, error) {
	tables, err := s.tableService.GetTablesBase(&dto.FilterTablesRequest{})
	if err != nil {
		return nil, err
	}

	orgMap := make(map[string]*struct {
		name      string
		total     int
		completed int
	})

	for _, table := range tables {
		if table.Organization == nil {
			continue
		}

		orgID := *table.OrganizationID
		if _, exists := orgMap[orgID]; !exists {
			orgMap[orgID] = &struct {
				name      string
				total     int
				completed int
			}{
				name: table.Organization.Name,
			}
		}

		orgMap[orgID].total++
		if table.Status == "finalized" || table.Status == "submitted" {
			orgMap[orgID].completed++
		}
	}

	var result []dto.OrganizationCompletionResponse
	for _, data := range orgMap {
		completion := 0.0
		if data.total > 0 {
			completion = (float64(data.completed) / float64(data.total)) * 100
		}

		result = append(result, dto.OrganizationCompletionResponse{
			Name:       data.name,
			Completion: completion,
			Tables:     data.total,
		})
	}

	// Sort by completion descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].Completion > result[j].Completion
	})

	return result, nil
}

func (s *DashboardService) GetTopPerformers() ([]dto.TopPerformerResponse, error) {
	tables, err := s.tableService.GetTablesBase(&dto.FilterTablesRequest{})
	if err != nil {
		return nil, err
	}

	orgMap := make(map[string]*struct {
		name           string
		total          int
		completed      int
		totalDays      float64
		completedCount int
	})

	for _, table := range tables {
		if table.Organization == nil {
			continue
		}

		orgID := *table.OrganizationID
		if _, exists := orgMap[orgID]; !exists {
			orgMap[orgID] = &struct {
				name           string
				total          int
				completed      int
				totalDays      float64
				completedCount int
			}{
				name: table.Organization.Name,
			}
		}

		orgMap[orgID].total++
		if table.Status == "finalized" || table.Status == "submitted" {
			orgMap[orgID].completed++

			// Calculate days from collection start date to table completion
			completionDate := table.UpdatedAt
			if completionDate.Before(CollectionStartDate) {
				// If updated before collection start, use collection start
				completionDate = CollectionStartDate
			}

			days := completionDate.Sub(CollectionStartDate).Hours() / 24
			if days >= 0 {
				orgMap[orgID].totalDays += days
				orgMap[orgID].completedCount++
			}
		}
	}

	var result []dto.TopPerformerResponse
	for _, data := range orgMap {
		if data.completedCount == 0 {
			continue
		}

		completion := (float64(data.completed) / float64(data.total)) * 100
		avgDays := data.totalDays / float64(data.completedCount)

		avgTime := ""
		if avgDays < 1 {
			avgTime = "< 1 day"
		} else {
			avgTime = formatDays(avgDays)
		}

		result = append(result, dto.TopPerformerResponse{
			Name:       data.name,
			AvgTime:    avgTime,
			Completion: completion,
		})
	}

	// Sort by combination of speed and completion
	sort.Slice(result, func(i, j int) bool {
		// Higher completion and lower avg time is better
		return result[i].Completion > result[j].Completion
	})

	// Assign ranks
	for i := range result {
		result[i].Rank = i + 1
	}

	// Return top 3
	if len(result) > 3 {
		result = result[:3]
	}

	return result, nil
}

func (s *DashboardService) GetOrganizationsNeedAttention() ([]dto.OrganizationNeedAttentionResponse, error) {
	tables, err := s.tableService.GetTablesBase(&dto.FilterTablesRequest{})
	if err != nil {
		return nil, err
	}

	orgMap := make(map[string]*struct {
		name        string
		total       int
		completed   int
		lastUpdated time.Time
	})

	for _, table := range tables {
		if table.Organization == nil {
			continue
		}

		orgID := *table.OrganizationID
		if _, exists := orgMap[orgID]; !exists {
			orgMap[orgID] = &struct {
				name        string
				total       int
				completed   int
				lastUpdated time.Time
			}{
				name:        table.Organization.Name,
				lastUpdated: table.UpdatedAt,
			}
		}

		orgMap[orgID].total++
		if table.Status == "finalized" || table.Status == "submitted" {
			orgMap[orgID].completed++
		}

		// Track the most recent update
		if table.UpdatedAt.After(orgMap[orgID].lastUpdated) {
			orgMap[orgID].lastUpdated = table.UpdatedAt
		}
	}

	var result []dto.OrganizationNeedAttentionResponse
	now := time.Now()

	for _, data := range orgMap {
		completion := 0.0
		if data.total > 0 {
			completion = (float64(data.completed) / float64(data.total)) * 100
		}

		// Only include organizations with completion < 50%
		if completion >= 50 {
			continue
		}

		// Calculate idle days from last update or collection start
		referenceDate := data.lastUpdated
		if referenceDate.Before(CollectionStartDate) {
			referenceDate = CollectionStartDate
		}
		daysIdle := int(now.Sub(referenceDate).Hours() / 24)

		status := "Rendah"
		if completion < 30 || daysIdle > 5 {
			status = "Kritis"
		}

		result = append(result, dto.OrganizationNeedAttentionResponse{
			Name:       data.name,
			Completion: completion,
			Tables:     data.total,
			Status:     status,
			DaysIdle:   daysIdle,
		})
	}

	// Sort by completion ascending (lowest first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Completion < result[j].Completion
	})

	return result, nil
}

func formatDays(days float64) string {
	if days < 1 {
		return "< 1 day"
	}
	return fmt.Sprintf("%.1f days", days)
}
