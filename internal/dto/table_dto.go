package dto

type TableResponse struct {
	ID                 string                `json:"id"`
	Name               string                `json:"name"`
	Direction          int                   `json:"direction"`
	Description        *string               `json:"description,omitempty"`
	Indicator          *IndicatorResponse    `json:"indicator"`
	Organization       *OrganizationResponse `json:"organization"`
	Labels             []string              `json:"labels"`
	Notes              *string               `json:"notes"`
	IsLocked           bool                  `json:"is_locked"`
	Status             string                `json:"status"`
	Aggregate          *string               `json:"aggregate"`
	HasParentDimension bool                  `json:"has_parent_dimension"`
	Dimensions         []DimensionResponse   `json:"dimensions"`
	Facts              []FactResponse        `json:"facts"`
}

type TableListResponse struct {
	ID                  string                  `json:"id"`
	Name                string                  `json:"name"`
	Indicator           *IndicatorListResponse  `json:"indicator"`
	Organization        *OrganizationResponse   `json:"organization"`
	Labels              []string                `json:"labels"`
	Notes               *string                 `json:"notes"`
	IsLocked            bool                    `json:"is_locked"`
	IsAggregated        bool                    `json:"is_aggregated"`
	Status              string                  `json:"status"`
	HasParentDimension  bool                    `json:"has_parent_dimension"`
	Dimensions          []DimensionListResponse `json:"dimensions"`
	InsightFactsSummary *SummaryInsightFacts    `json:"insight_facts_summary"`
}

type CreateTableRequest struct {
	Name           string   `json:"name" validate:"required"`
	IndicatorID    string   `json:"indicator_id" validate:"required"`
	OrganizationID *string  `json:"organization_id"`
	DimensionIDs   []string `json:"dimension_ids" validate:"required,min=0,max=2,dive,required"`
}

type UpdateTableRequest struct {
	Name           *string  `json:"name,omitempty"`
	IndicatorID    *string  `json:"indicator_id,omitempty"`
	OrganizationID *string  `json:"organization_id,omitempty"`
	DimensionIDs   []string `json:"dimension_ids,omitempty"`
}

type AddLabelsToTablesRequest struct {
	Labels   []string `json:"labels" validate:"required,min=1,dive,required"`
	TableIDs []string `json:"table_ids" validate:"required,min=1,dive,required"`
}

type TableLabelResponse struct {
	Name string `json:"name"`
}

type UpdateTableLabelsRequest struct {
	Labels []string `json:"labels" validate:"required,min=0,dive,required"`
}

type UpdateTableNameRequest struct {
	Name string `json:"name" validate:"required"`
}

type UpdateTableNotesRequest struct {
	Notes *string `json:"notes" validate:"required"`
}

type UpdateTableIsLockedRequest struct {
	Locked bool `json:"locked" validate:"required"`
}

type UpdateTableStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft submitted finalized unfinalized"`
}

type AnalyzeTablesRequest struct {
	TableIDs []string `json:"table_ids" validate:"required,min=1,dive,required"`
}

type CommitTablesRequest struct {
	TableIDs []string `json:"table_ids" validate:"required,min=1,dive,required"`
}

type FilterTablesRequest struct {
	OrganizationID *string `json:"organization_id,omitempty"`
}

type TableExportResponse struct {
	Name string `json:"name"`
	File []byte `json:"file"`
}

type GenerateParentTableRequest struct {
	DimensionIDs []string `json:"dimension_ids" binding:"required,min=1"` // Dimension IDs yang akan diagregasi (bisa lebih dari 1)
}

// GenerateParentTableResponse adalah response setelah generate parent table
type GenerateParentTableResponse struct {
	ParentTableID        string                    `json:"parent_table_id"`
	IsNewTable           bool                      `json:"is_new_table"` // true jika tabel baru dibuat, false jika update existing
	Message              string                    `json:"message"`
	ChildTableID         string                    `json:"child_table_id"`
	AggregatedDimensions []AggregatedDimensionInfo `json:"aggregated_dimensions"` // Info dimensions yang diagregasi
}

// AggregatedDimensionInfo adalah informasi dimension yang diagregasi
type AggregatedDimensionInfo struct {
	DimensionID      string `json:"dimension_id"`
	DimensionName    string `json:"dimension_name"`
	ParentValuesUsed int    `json:"parent_values_used"` // Jumlah parent values yang digunakan
}
