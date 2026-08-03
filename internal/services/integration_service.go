package services

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"statio/config"
	"statio/internal/dto"
	"strconv"
	"strings"
	"time"
)

const (
	CONFIGURATION_KEY_WEBSITE_COOKIE    = "integration_website_cookie"
	PATH_INTEGRATION_UPLOAD_DATA        = "/backend/css/modelGetData/uploadData.php"
	PATH_INTEGRATION_SAVE_DATA          = "/backend/css/modelGetData/saveData.php"
	PATH_INTEGRATION_GET_VARIABLE       = "/backend/css/modelGetData/getVariabel.php"
	PATH_INTEGRATION_GET_VERTICAL       = "/backend/css/modelGetData/getVerticalVariabel.php"
	PATH_INTEGRATION_GET_CLASSIFICATION = "/backend/css/modelGetData/getKlasifikasiVariabel.php"
	PATH_INTEGRATION_GET_PERIOD         = "/backend/css/modelGetData/getPeriodeWaktu.php"
)

type IntegrationService struct {
	tableSvc          *TableService
	configSvc         *ConfigurationService
	integrationConfig *config.IntegrationConfig
}

func NewIntegrationService(
	tableSvc *TableService,
	configSvc *ConfigurationService,
	integrationConfig *config.IntegrationConfig,
) *IntegrationService {
	return &IntegrationService{
		tableSvc:          tableSvc,
		configSvc:         configSvc,
		integrationConfig: integrationConfig,
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
	var metadata []dto.ExportMetadata

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
		metadataEntry := dto.ExportMetadata{
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

func (s *IntegrationService) ImportDataIntegration(file *multipart.FileHeader) (
	*dto.IntegrationUploadResponse, error,
) {
	if file == nil {
		return nil, fmt.Errorf("file is required")
	}

	archiveData, err := s.readMultipartFile(file)
	if err != nil {
		return nil, err
	}

	items, zipReader, err := s.readZipMetadata(archiveData)
	if err != nil {
		return nil, err
	}

	baseURL := s.integrationConfig.WebsiteBaseURL
	referer := s.integrationConfig.WebsiteReferer
	cookieConfig, err := s.configSvc.GetConfigurationByKey(CONFIGURATION_KEY_WEBSITE_COOKIE)
	if err != nil {
		return nil, fmt.Errorf("failed to get integration cookie configuration: %w", err)
	}

	cookie := ""
	if cookieConfig != nil {
		cookie = strings.TrimSpace(cookieConfig.Value)
	}

	if strings.TrimSpace(cookie) == "" {
		return nil, fmt.Errorf("integration website cookie is not configured")
	}

	results := make([]dto.IntegrationUploadResult, 0)
	errors := make([]dto.IntegrationUploadError, 0)

	for _, item := range items {
		entryName := fmt.Sprintf("files/%s", item.File)
		entry := zipReader.File[0]

		found := false
		for _, f := range zipReader.File {
			if f.Name == entryName {
				entry = f
				found = true
				break
			}
		}

		if !found {
			errors = append(errors, dto.IntegrationUploadError{
				Item:  item,
				Error: fmt.Sprintf("File not found: %s", entryName),
			})
			continue
		}

		fileData, readErr := s.readZipEntry(entry)
		if readErr != nil {
			errors = append(errors, dto.IntegrationUploadError{
				Item:  item,
				Error: fmt.Sprintf("Failed reading file %s: %s", entryName, readErr.Error()),
			})
			continue
		}

		payload, payloadErr := s.generatePayload(item, fileData, baseURL, referer, cookie)
		if payloadErr != nil {
			errors = append(errors, dto.IntegrationUploadError{
				Item:  item,
				Error: payloadErr.Error(),
			})
			continue
		}

		uploadErr := s.postMultipart(baseURL+PATH_INTEGRATION_UPLOAD_DATA, payload.UploadFields, fileData, item.File, cookie, referer)
		if uploadErr != nil {
			errors = append(errors, dto.IntegrationUploadError{
				Item:  item,
				Error: fmt.Sprintf("Upload failed: %s", uploadErr.Error()),
			})
			continue
		}

		saveResp, saveErr := s.postMultipartWithJSONResponse(baseURL+PATH_INTEGRATION_SAVE_DATA, payload.SaveFields, fileData, item.File, cookie, referer)
		if saveErr != nil {
			errors = append(errors, dto.IntegrationUploadError{
				Item:  item,
				Error: fmt.Sprintf("Save failed: %s", saveErr.Error()),
			})
			continue
		}

		results = append(results, dto.IntegrationUploadResult{
			Item:     item,
			Success:  true,
			Response: saveResp,
		})
	}

	return &dto.IntegrationUploadResponse{
		Message: "Processing complete",
		Total:   len(items),
		Success: len(results),
		Failed:  len(errors),
		Results: results,
		Errors:  errors,
	}, nil
}

type integrationFormPayload struct {
	UploadFields map[string]string
	SaveFields   map[string]string
}

func (s *IntegrationService) generatePayload(item dto.ExportMetadata, fileData []byte, baseURL, referer, cookie string) (*integrationFormPayload, error) {
	variables, err := s.getVariables(item.SubjectID, baseURL, referer, cookie)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch variables for subject %s: %w", item.SubjectID, err)
	}

	var variable *dto.OptionWebItem
	for i := range variables {
		if variables[i].Value == item.ID {
			variable = &variables[i]
			break
		}
	}

	if variable == nil {
		return nil, fmt.Errorf("variable with id %s not found for subject %s", item.ID, item.SubjectID)
	}

	verticalItems, verticalLabel, err := s.getVerticalData(item.ID, baseURL, referer, cookie)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch vertical for variable %s: %w", item.ID, err)
	}

	classificationItems, err := s.getClassificationData(item.ID, baseURL, referer, cookie)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch classification for variable %s: %w", item.ID, err)
	}

	periodItems, yearOptions, err := s.getPeriodData(item.ID, baseURL, referer, cookie)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch period for variable %s: %w", item.ID, err)
	}

	hasClassification := len(classificationItems) > 0
	hasPeriod := len(periodItems) > 0

	var validYear *dto.OptionWebItem
	yearLabel := strconv.Itoa(item.Year)
	for i := range yearOptions {
		if yearOptions[i].Label == yearLabel {
			validYear = &yearOptions[i]
			break
		}
	}

	if validYear == nil {
		return nil, fmt.Errorf("year %d not found for variable %s", item.Year, item.ID)
	}

	formatTable := s.getFormatTable(hasPeriod, hasClassification)
	if formatTable == "" {
		return nil, fmt.Errorf("failed to determine format table for variable %s", item.ID)
	}

	itemVerticalVar := toOptionItems(verticalItems)
	itemKlasifikasiVar := toOptionItems(classificationItems)
	itemTurunanTahun := toOptionItems(periodItems)
	tahunObj := dto.OptionItem{Nama: validYear.Label, Value: validYear.Value}
	tahunArr := []dto.OptionItem{tahunObj}

	namaKelVerVar := "( " + verticalLabel + " )"

	uploadFields := map[string]string{
		"itemVerticalVar": s.mustJSON(itemVerticalVar),
		"namaVariabel":    variable.Label,
		"namaKelVerVar":   namaKelVerVar,
		"formatTabel":     formatTable,
	}

	saveFields := map[string]string{
		"itemVerticalVar": s.mustJSON(itemVerticalVar),
		"namaVariabel":    variable.Label,
		"idVariabel":      variable.Value,
		"formatSaveData":  formatTable,
		"eksekusi":        "deleteInsert",
	}

	switch formatTable {
	case "1":
		uploadFields["itemKlasifikasiVar"] = s.mustJSON(itemKlasifikasiVar)
		uploadFields["itemTurunanTahun"] = s.mustJSON(itemTurunanTahun)
		uploadFields["tahun"] = s.mustJSON(tahunObj)

		saveFields["itemKlasifikasiVar"] = s.mustJSON(itemKlasifikasiVar)
		saveFields["itemTurunanTahun"] = s.mustJSON(itemTurunanTahun)
		saveFields["tahun"] = s.mustJSON(tahunObj)
		saveFields["tahun_duplicate"] = s.mustJSON(tahunObj)
	case "2":
		uploadFields["itemKlasifikasiVar"] = s.mustJSON(itemKlasifikasiVar)
		uploadFields["tahun"] = s.mustJSON(tahunArr)

		saveFields["itemKlasifikasiVar"] = s.mustJSON(itemKlasifikasiVar)
		saveFields["tahun"] = s.mustJSON(tahunArr)
		saveFields["tahun_duplicate"] = s.mustJSON(tahunArr)
	case "3":
		uploadFields["itemTurunanTahun"] = s.mustJSON(itemTurunanTahun)
		uploadFields["tahun"] = s.mustJSON(tahunObj)

		saveFields["itemTurunanTahun"] = s.mustJSON(itemTurunanTahun)
		saveFields["tahun"] = s.mustJSON(tahunObj)
		saveFields["tahun_duplicate"] = s.mustJSON(tahunObj)
	default:
		uploadFields["tahun"] = s.mustJSON(tahunArr)
		saveFields["tahun"] = s.mustJSON(tahunArr)
		saveFields["tahun_duplicate"] = s.mustJSON(tahunArr)
	}

	_ = fileData

	return &integrationFormPayload{
		UploadFields: uploadFields,
		SaveFields:   saveFields,
	}, nil
}

func (s *IntegrationService) getFormatTable(hasPeriod, hasClassification bool) string {
	if hasPeriod && hasClassification {
		return "1"
	}
	if !hasPeriod && hasClassification {
		return "2"
	}
	if hasPeriod && !hasClassification {
		return "3"
	}
	return "4"
}

func (s *IntegrationService) readMultipartFile(file *multipart.FileHeader) ([]byte, error) {
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read uploaded file: %w", err)
	}

	return data, nil
}

func (s *IntegrationService) readZipMetadata(archiveData []byte) ([]dto.ExportMetadata, *zip.Reader, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(archiveData), int64(len(archiveData)))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read zip file: %w", err)
	}

	var metadataFile *zip.File
	for _, f := range zipReader.File {
		if f.Name == "metadata.json" {
			metadataFile = f
			break
		}
	}

	if metadataFile == nil {
		return nil, nil, fmt.Errorf("metadata.json not found in zip")
	}

	metadataData, err := s.readZipEntry(metadataFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read metadata.json: %w", err)
	}

	var metadata []dto.ExportMetadata
	if err := json.Unmarshal(metadataData, &metadata); err != nil {
		return nil, nil, fmt.Errorf("failed to parse metadata.json: %w", err)
	}

	return metadata, zipReader, nil
}

func (s *IntegrationService) readZipEntry(file *zip.File) ([]byte, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return io.ReadAll(rc)
}

func (s *IntegrationService) getVariables(subjectID, baseURL, referer, cookie string) ([]dto.OptionWebItem, error) {
	body := "id_subject=" + url.QueryEscape(subjectID)
	respData, err := s.postURLEncoded(baseURL+PATH_INTEGRATION_GET_VARIABLE, body, cookie, referer)
	if err != nil {
		return nil, err
	}

	var items []dto.OptionWebItem
	if err := json.Unmarshal(respData, &items); err != nil {
		return nil, err
	}

	return items, nil
}

func (s *IntegrationService) getVerticalData(variableID, baseURL, referer, cookie string) ([]dto.OptionWebItem, string, error) {
	body := "id_variabel=" + url.QueryEscape(variableID)
	respData, err := s.postURLEncoded(baseURL+PATH_INTEGRATION_GET_VERTICAL, body, cookie, referer)
	if err != nil {
		return nil, "", err
	}

	var arr []json.RawMessage
	if err := json.Unmarshal(respData, &arr); err != nil {
		return nil, "", err
	}

	if len(arr) < 5 {
		return nil, "", fmt.Errorf("unexpected vertical response shape")
	}

	items, err := parseOptionWebItems(arr[1])
	if err != nil {
		return nil, "", err
	}

	var verticalLabel string
	if err := json.Unmarshal(arr[4], &verticalLabel); err != nil {
		verticalLabel = ""
	}

	return items, verticalLabel, nil
}

func (s *IntegrationService) getClassificationData(variableID, baseURL, referer, cookie string) ([]dto.OptionWebItem, error) {
	body := "id_variabel=" + url.QueryEscape(variableID)
	respData, err := s.postURLEncoded(baseURL+PATH_INTEGRATION_GET_CLASSIFICATION, body, cookie, referer)
	if err != nil {
		return nil, err
	}

	var arr []json.RawMessage
	if err := json.Unmarshal(respData, &arr); err != nil {
		return nil, err
	}

	if len(arr) < 2 {
		return nil, fmt.Errorf("unexpected classification response shape")
	}

	items, err := parseOptionWebItems(arr[1])
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *IntegrationService) getPeriodData(variableID, baseURL, referer, cookie string) ([]dto.OptionWebItem, []dto.OptionWebItem, error) {
	body := "id_variabel=" + url.QueryEscape(variableID)
	respData, err := s.postURLEncoded(baseURL+PATH_INTEGRATION_GET_PERIOD, body, cookie, referer)
	if err != nil {
		return nil, nil, err
	}

	var arr []json.RawMessage
	if err := json.Unmarshal(respData, &arr); err != nil {
		return nil, nil, err
	}

	if len(arr) < 4 {
		return nil, nil, fmt.Errorf("unexpected period response shape")
	}

	periodItems, err := parseOptionWebItems(arr[1])
	if err != nil {
		return nil, nil, err
	}

	var years []dto.OptionWebItem
	if err := json.Unmarshal(arr[3], &years); err != nil {
		return nil, nil, err
	}

	return periodItems, years, nil
}

func (s *IntegrationService) postURLEncoded(endpoint, payload, cookie, referer string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Referer", referer)

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (s *IntegrationService) postMultipart(endpoint string, fields map[string]string, fileData []byte, filename, cookie, referer string) error {
	body, contentType, err := s.buildMultipartBody(fields, fileData, filename)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Referer", referer)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (s *IntegrationService) postMultipartWithJSONResponse(endpoint string, fields map[string]string, fileData []byte, filename, cookie, referer string) (any, error) {
	body, contentType, err := s.buildMultipartBody(fields, fileData, filename)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Referer", referer)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var data any
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return map[string]any{"raw": base64.StdEncoding.EncodeToString(bodyBytes)}, nil
	}

	return data, nil
}

func (s *IntegrationService) buildMultipartBody(fields map[string]string, fileData []byte, filename string) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	filePart, err := writer.CreateFormFile("bttUploadData", filename)
	if err != nil {
		return nil, "", err
	}

	if _, err := filePart.Write(fileData); err != nil {
		return nil, "", err
	}

	for key, value := range fields {
		if key == "tahun_duplicate" {
			if err := writer.WriteField("tahun", value); err != nil {
				return nil, "", err
			}
			continue
		}

		if err := writer.WriteField(key, value); err != nil {
			return nil, "", err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}

func parseOptionWebItems(raw json.RawMessage) ([]dto.OptionWebItem, error) {
	if string(raw) == "null" {
		return []dto.OptionWebItem{}, nil
	}

	var items []dto.OptionWebItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}

	return items, nil
}

func toOptionItems(items []dto.OptionWebItem) []dto.OptionItem {
	result := make([]dto.OptionItem, 0, len(items))
	for _, item := range items {
		result = append(result, dto.OptionItem{
			Nama:  item.Label,
			Value: item.Value,
		})
	}
	return result
}

func (s *IntegrationService) mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(b)
}
