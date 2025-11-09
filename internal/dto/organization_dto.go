package dto

type OrganizationResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AssignTablesRequest struct {
	TableIDs []string `json:"table_ids" validate:"required,min=1,dive,required"`
}
