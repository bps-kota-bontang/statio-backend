package services

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"statio/internal/dto"
	"time"
)

const (
	CONFIGURATION_KEY_BASE_URL = "integration_base_url"
	PATH_INTEGRATION_UPLOAD    = "/api/v1/integrations/upload"
)

type IntegrationService struct {
	tableSvc  *TableService
	configSvc *ConfigurationService
}

type ExportMetadata struct {
	ID        string `json:"id"`
	SubjectID string `json:"subjectId"`
	Year      int    `json:"year"`
	File      string `json:"file"`
}

func NewIntegrationService(
	tableSvc *TableService,
	configSvc *ConfigurationService,
) *IntegrationService {
	return &IntegrationService{
		tableSvc:  tableSvc,
		configSvc: configSvc,
	}
}

func (s *IntegrationService) ExportDataIntegration(tableIDs []string, year int) (*dto.FileResponse, error) {
	tables, err := s.tableSvc.GetTablesBase(
		&dto.FilterTablesRequest{
			TableIDs: tableIDs,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find tables: %w", err)
	}

	// Create a buffer to write our archive to
	buf := new(bytes.Buffer)

	// Create a new zip archive
	zipWriter := zip.NewWriter(buf)
	defer zipWriter.Close()

	// Slice to store metadata
	var metadata []ExportMetadata

	// Download each table
	for _, table := range tables {
		// Skip tables without website IDs
		if table.WebsiteTableID == nil || table.WebsiteSubjectID == nil {
			continue
		}

		// Download table to file
		downloadData, err := s.tableSvc.DownloadTable(table.ID, []int{year}, "xls", []string{"admin"}, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to download table %s: %w", table.ID, err)
		}

		// Add file to zip under "files" folder
		filename := fmt.Sprintf("files/%s", downloadData.Name)
		fileWriter, err := zipWriter.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to create zip entry for %s: %w", filename, err)
		}

		_, err = fileWriter.Write(downloadData.File)
		if err != nil {
			return nil, fmt.Errorf("failed to write file %s to zip: %w", filename, err)
		}

		// Add metadata entry
		metadataEntry := ExportMetadata{
			ID:        *table.WebsiteTableID,
			SubjectID: *table.WebsiteSubjectID,
			Year:      year,
			File:      downloadData.Name,
		}

		metadata = append(metadata, metadataEntry)
	}

	// Check if any valid tables were exported
	if len(metadata) == 0 {
		return nil, fmt.Errorf("no tables with website source found to export")
	}

	// Create metadata.json
	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Add metadata.json to zip root
	metadataWriter, err := zipWriter.Create("metadata.json")
	if err != nil {
		return nil, fmt.Errorf("failed to create metadata.json in zip: %w", err)
	}

	_, err = metadataWriter.Write(metadataJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to write metadata.json: %w", err)
	}

	// Close the zip writer to flush all data
	err = zipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("data_integration_export_%s.zip", timestamp)

	return &dto.FileResponse{
		Name: filename,
		File: buf.Bytes(),
	}, nil
}

func (s *IntegrationService) ImportDataIntegration(file *multipart.FileHeader) error {
	configBaseURLResp, err := s.configSvc.GetConfigurationByKey(CONFIGURATION_KEY_BASE_URL)
	if err != nil {
		return fmt.Errorf("failed to get integration base URL configuration: %w", err)
	}

	if configBaseURLResp == nil || configBaseURLResp.Value == "" {
		return fmt.Errorf("integration base URL configuration is not set")
	}

	baseURL := configBaseURLResp.Value

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create a buffer to store the multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a form file field
	part, err := writer.CreateFormFile("file", file.Filename)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy the file content to the form field
	_, err = io.Copy(part, src)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Close the multipart writer to finalize the form
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create the full upload URL
	uploadURL := baseURL + PATH_INTEGRATION_UPLOAD

	// Create HTTP request
	req, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set the content type with the boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send upload request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
