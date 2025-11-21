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

type SummaryMissingFacts struct {
	ExpectedPerYear int `json:"expected_per_year"`
	TotalExpected   int `json:"total_expected"`
	TotalFilled     int `json:"total_filled"`
	TotalMissing    int `json:"total_missing"`
}

type DataMissingFact struct {
	Year     int `json:"year"`
	Expected int `json:"expected"`
	Filled   int `json:"filled"`
	Missing  int `json:"missing"`
}

type MissingFactsResponse struct {
	TableID  string              `json:"table_id"`
	FromYear int                 `json:"from_year"`
	ToYear   int                 `json:"to_year"`
	Summary  SummaryMissingFacts `json:"summary"`
	Data     []DataMissingFact   `json:"data"`
}

type AnalyzeFactPayload struct {
	TableID string `json:"table_id"`
}
