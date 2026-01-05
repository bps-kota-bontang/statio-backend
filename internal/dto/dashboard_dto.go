package dto

type DashboardStatisticsResponse struct {
	TotalTables         int `json:"total_tables"`
	TotalTableDraft     int `json:"total_table_draft"`
	TotalTableSubmitted int `json:"total_table_submitted"`
	TotalTableFinalized int `json:"total_table_finalized"`
}
