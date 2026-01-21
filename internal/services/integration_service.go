package services

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"statio/internal/dto"
	"time"
)

type IntegrationService struct {
	tableSvc *TableService
}

type ExportMetadata struct {
	ID        string `json:"id"`
	SubjectID string `json:"subjectId"`
	Year      int    `json:"year"`
	File      string `json:"file"`
}

func NewIntegrationService(tableSvc *TableService) *IntegrationService {
	return &IntegrationService{
		tableSvc: tableSvc,
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

	// Export each table
	for _, table := range tables {
		// Skip tables without website IDs
		if table.WebsiteTableID == nil || table.WebsiteSubjectID == nil {
			continue
		}

		// Export table to file
		exportData, err := s.tableSvc.ExportTable(table.ID, []int{year}, "xls")
		if err != nil {
			return nil, fmt.Errorf("failed to export table %s: %w", table.ID, err)
		}

		// Add file to zip under "files" folder
		filename := fmt.Sprintf("files/%s", exportData.Name)
		fileWriter, err := zipWriter.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to create zip entry for %s: %w", filename, err)
		}

		_, err = fileWriter.Write(exportData.File)
		if err != nil {
			return nil, fmt.Errorf("failed to write file %s to zip: %w", filename, err)
		}

		// Add metadata entry
		metadataEntry := ExportMetadata{
			ID:        *table.WebsiteTableID,
			SubjectID: *table.WebsiteSubjectID,
			Year:      year,
			File:      exportData.Name,
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
