package dto

type DashboardStatisticsResponse struct {
	TotalTables         int `json:"total_tables"`
	TotalTableDraft     int `json:"total_table_draft"`
	TotalTableSubmitted int `json:"total_table_submitted"`
	TotalTableFinalized int `json:"total_table_finalized"`
}

type OrganizationCompletionResponse struct {
	Name       string  `json:"name"`
	Completion float64 `json:"completion"`
	Tables     int     `json:"tables"`
}

type TopPerformerResponse struct {
	Name       string  `json:"name"`
	AvgTime    string  `json:"avg_time"`
	Completion float64 `json:"completion"`
	Rank       int     `json:"rank"`
}

type OrganizationNeedAttentionResponse struct {
	Name       string  `json:"name"`
	Completion float64 `json:"completion"`
	Tables     int     `json:"tables"`
	Status     string  `json:"status"`
	DaysIdle   int     `json:"days_idle"`
}
