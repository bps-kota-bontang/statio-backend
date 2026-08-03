package dto

type ProjectListResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	TableCount int    `json:"table_count"`
}

type ProjectTableResponse struct {
	TableID     string `json:"table_id"`
	TableName   string `json:"table_name"`
	Page        string `json:"page"`
	TableNumber string `json:"table_number"`
}

type ProjectResponse struct {
	ID     string                 `json:"id"`
	Name   string                 `json:"name"`
	Tables []ProjectTableResponse `json:"tables"`
}

type CreateProjectRequest struct {
	Name string `json:"name" validate:"required"`
}

type UpdateProjectRequest struct {
	Name string `json:"name" validate:"required"`
}

type ProjectTableRequest struct {
	TableID     string `json:"table_id" validate:"required"`
	Page        string `json:"page" validate:"required"`
	TableNumber string `json:"table_number" validate:"required"`
}
