package dto

type FactDimensionResponse struct {
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Value DimensionValueResponse `json:"value"`
}

type FactResponse struct {
	OldValue   *float64                `json:"old_value"`
	Value      *float64                `json:"value"`
	Year       int                     `json:"year"`
	IsOutlier  *bool                   `json:"is_outlier"`
	Dimensions []FactDimensionResponse `json:"dimensions"`
}

type FactPayload struct {
	Dimensions []string `json:"dimensions" validate:"required,min=0,max=2,dive,required"`
	Value      *float64 `json:"value"`
	Year       int      `json:"year"`
}

type UpdateFactRequest struct {
	Data []FactPayload `json:"data"`
}

type SummaryInsightFacts struct {
	ExpectedPerYear int `json:"expected_per_year"`
	TotalExpecteds  int `json:"total_expecteds"`
	TotalFilleds    int `json:"total_filleds"`
	TotalMissings   int `json:"total_missings"`
	TotalRevisions  int `json:"total_revisions"`
	TotalOutliers   int `json:"total_outliers"`
}

type DataInsightFact struct {
	Year     int `json:"year"`
	Expected int `json:"expected"`
	Filled   int `json:"filled"`
	Missing  int `json:"missing"`
	Revision int `json:"revision"`
	Outlier  int `json:"outlier"`
}

type InsightFactsResponse struct {
	TableID  string              `json:"table_id"`
	FromYear int                 `json:"from_year"`
	ToYear   int                 `json:"to_year"`
	Summary  SummaryInsightFacts `json:"summary"`
	Data     []DataInsightFact   `json:"data"`
}

type UnanalyzeFactPayload struct {
	TableID string `json:"table_id"`
}

type AnalyzeFactPayload struct {
	TableID string `json:"table_id"`
}
