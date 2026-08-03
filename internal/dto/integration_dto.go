package dto

type ExportMetadata struct {
	ID        string `json:"id"`
	SubjectID string `json:"subjectId"`
	Year      int    `json:"year"`
	File      string `json:"file"`
}

type OptionWebItem struct {
	HTML  string `json:"html"`
	Label string `json:"label"`
	Value string `json:"value"`
}

type OptionItem struct {
	Nama  string `json:"nama"`
	Value string `json:"value"`
}

type IntegrationUploadResult struct {
	Item     ExportMetadata `json:"item"`
	Success  bool           `json:"success"`
	Response any            `json:"response,omitempty"`
}

type IntegrationUploadError struct {
	Item  ExportMetadata `json:"item"`
	Error string         `json:"error"`
}

type IntegrationUploadResponse struct {
	Message string                    `json:"message"`
	Total   int                       `json:"total"`
	Success int                       `json:"success"`
	Failed  int                       `json:"failed"`
	Results []IntegrationUploadResult `json:"results"`
	Errors  []IntegrationUploadError  `json:"errors"`
}
