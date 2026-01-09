package services

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"statio/internal/dto"
	"strings"

	"github.com/xuri/excelize/v2"
)

type ExcelService struct{}

func NewExcelService() *ExcelService {
	return &ExcelService{}
}

// ExportToXLSX exports table data to Excel .xlsx format
func (s *ExcelService) ExportToXLSX(table *dto.TableResponse, years []int) ([]byte, error) {
	// Create a new Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("failed to close excel file: %v", err)
		}
	}()

	// Sort years
	sort.Ints(years)

	// Create styles
	headerStyle, dataStyle, err := s.createStyles(f)
	if err != nil {
		return nil, err
	}

	// Generate sheets based on dimension count
	switch len(table.Dimensions) {
	case 0:
		if err := s.exportDimension0(f, table, years, headerStyle, dataStyle); err != nil {
			return nil, err
		}
	case 1:
		if err := s.exportDimension1(f, table, years, headerStyle, dataStyle); err != nil {
			return nil, err
		}
	case 2:
		if err := s.exportDimension2(f, table, years, headerStyle, dataStyle); err != nil {
			return nil, err
		}
	default:
		if err := s.exportDimension3Plus(f, table, years, headerStyle, dataStyle); err != nil {
			return nil, err
		}
	}

	// Save to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write excel file: %w", err)
	}

	return buf.Bytes(), nil
}

// ExportToXLS exports table data to Excel .xls format (legacy)
func (s *ExcelService) ExportToXLS(table *dto.TableResponse, years []int) ([]byte, error) {
	// For now, export as XLSX - browsers can still open it
	// TODO: Implement true .xls format using a different library if needed
	return s.ExportToXLSX(table, years)
}

// createStyles creates header and data cell styles
func (s *ExcelService) createStyles(f *excelize.File) (int, int, error) {
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#D3D3D3"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		return 0, 0, err
	}

	dataStyle, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		return 0, 0, err
	}

	return headerStyle, dataStyle, nil
}

// exportDimension0 exports table with 0 dimensions (organization-only)
func (s *ExcelService) exportDimension0(f *excelize.File, table *dto.TableResponse, years []int, headerStyle, dataStyle int) error {
	sheetName := "Sheet1"
	currentRow := 1

	// Header row 1: "Kabupaten/Kota" (A1:A2 merged) and "Tahun" (B1 to last column merged)
	cellA1, _ := excelize.CoordinatesToCellName(1, currentRow)
	cellA2, _ := excelize.CoordinatesToCellName(1, currentRow+1)
	f.SetCellValue(sheetName, cellA1, "Kabupaten/Kota")
	f.MergeCell(sheetName, cellA1, cellA2)
	f.SetCellStyle(sheetName, cellA1, cellA2, headerStyle)

	cellB1, _ := excelize.CoordinatesToCellName(2, currentRow)
	cellLast1, _ := excelize.CoordinatesToCellName(len(years)+1, currentRow)
	f.SetCellValue(sheetName, cellB1, "Tahun")
	f.MergeCell(sheetName, cellB1, cellLast1)
	f.SetCellStyle(sheetName, cellB1, cellLast1, headerStyle)
	currentRow++

	// Header row 2: year values
	for colIdx, year := range years {
		cellName, _ := excelize.CoordinatesToCellName(colIdx+2, currentRow)
		f.SetCellValue(sheetName, cellName, year)
		f.SetCellStyle(sheetName, cellName, cellName, headerStyle)
	}
	currentRow++

	// Data row: organization name and values
	yearValues := make(map[int]interface{})
	for _, fact := range table.Facts {
		if fact.Value != nil {
			yearValues[fact.Year] = *fact.Value
		}
	}

	cellA3, _ := excelize.CoordinatesToCellName(1, currentRow)
	f.SetCellValue(sheetName, cellA3, "Kota Bontang")
	f.SetCellStyle(sheetName, cellA3, cellA3, dataStyle)

	for colIdx, year := range years {
		cellName, _ := excelize.CoordinatesToCellName(colIdx+2, currentRow)
		if yearValues[year] != nil {
			f.SetCellValue(sheetName, cellName, yearValues[year])
		} else {
			f.SetCellValue(sheetName, cellName, "")
		}
		f.SetCellStyle(sheetName, cellName, cellName, dataStyle)
	}

	// Auto-fit columns
	f.SetColWidth(sheetName, "A", "A", 25)
	for i := range years {
		colName, _ := excelize.ColumnNumberToName(i + 2)
		f.SetColWidth(sheetName, colName, colName, 15)
	}

	return nil
}

// exportDimension1 exports table with 1 dimension
func (s *ExcelService) exportDimension1(f *excelize.File, table *dto.TableResponse, years []int, headerStyle, dataStyle int) error {
	sheetName := "Sheet1"
	currentRow := 1

	dim := table.Dimensions[0]

	// Collect dimension values
	dimValues := []string{}
	for _, dimValue := range dim.Values {
		dimValues = append(dimValues, dimValue.Name)
	}

	// Header row 1: dimension name (A1:A2 merged) and "Tahun" (B1 to last column merged)
	cellA1, _ := excelize.CoordinatesToCellName(1, currentRow)
	cellA2, _ := excelize.CoordinatesToCellName(1, currentRow+1)
	f.SetCellValue(sheetName, cellA1, dim.Name)
	f.MergeCell(sheetName, cellA1, cellA2)
	f.SetCellStyle(sheetName, cellA1, cellA2, headerStyle)

	cellB1, _ := excelize.CoordinatesToCellName(2, currentRow)
	cellLast1, _ := excelize.CoordinatesToCellName(len(years)+1, currentRow)
	f.SetCellValue(sheetName, cellB1, "Tahun")
	f.MergeCell(sheetName, cellB1, cellLast1)
	f.SetCellStyle(sheetName, cellB1, cellLast1, headerStyle)
	currentRow++

	// Header row 2: year values
	for colIdx, year := range years {
		cellName, _ := excelize.CoordinatesToCellName(colIdx+2, currentRow)
		f.SetCellValue(sheetName, cellName, year)
		f.SetCellStyle(sheetName, cellName, cellName, headerStyle)
	}
	currentRow++

	// Build pivot data
	pivotData := make(map[string]map[int]interface{})
	for _, fact := range table.Facts {
		var dimVal string
		for _, fv := range fact.Dimensions {
			if fv.ID == dim.ID {
				dimVal = fv.Value.Name
				break
			}
		}
		if pivotData[dimVal] == nil {
			pivotData[dimVal] = make(map[int]interface{})
		}
		if fact.Value != nil {
			pivotData[dimVal][fact.Year] = *fact.Value
		}
	}

	// Data rows
	for _, dimVal := range dimValues {
		cellName, _ := excelize.CoordinatesToCellName(1, currentRow)
		f.SetCellValue(sheetName, cellName, dimVal)
		f.SetCellStyle(sheetName, cellName, cellName, dataStyle)

		for colIdx, year := range years {
			cellName, _ := excelize.CoordinatesToCellName(colIdx+2, currentRow)
			if pivotData[dimVal] != nil && pivotData[dimVal][year] != nil {
				f.SetCellValue(sheetName, cellName, pivotData[dimVal][year])
			} else {
				f.SetCellValue(sheetName, cellName, "")
			}
			f.SetCellStyle(sheetName, cellName, cellName, dataStyle)
		}
		currentRow++
	}

	// Auto-fit columns
	f.SetColWidth(sheetName, "A", "A", 25)
	for i := range years {
		colName, _ := excelize.ColumnNumberToName(i + 2)
		f.SetColWidth(sheetName, colName, colName, 15)
	}

	return nil
}

// exportDimension2 exports table with 2 dimensions (multiple sheets per year)
func (s *ExcelService) exportDimension2(f *excelize.File, table *dto.TableResponse, years []int, headerStyle, dataStyle int) error {
	dim1 := table.Dimensions[0]
	dim2 := table.Dimensions[1]

	// Collect dimension values
	dim1Values := []string{}
	dim2Values := []string{}

	for _, dimValue := range dim1.Values {
		dim1Values = append(dim1Values, dimValue.Name)
	}

	for _, dimValue := range dim2.Values {
		dim2Values = append(dim2Values, dimValue.Name)
	}

	// Build pivot data by year
	pivotDataByYear := make(map[int]map[string]map[string]interface{})
	for _, fact := range table.Facts {
		var dim1Val, dim2Val string
		for _, fv := range fact.Dimensions {
			if fv.ID == dim1.ID {
				dim1Val = fv.Value.Name
			}
			if fv.ID == dim2.ID {
				dim2Val = fv.Value.Name
			}
		}
		if pivotDataByYear[fact.Year] == nil {
			pivotDataByYear[fact.Year] = make(map[string]map[string]interface{})
		}
		if pivotDataByYear[fact.Year][dim1Val] == nil {
			pivotDataByYear[fact.Year][dim1Val] = make(map[string]interface{})
		}
		if fact.Value != nil {
			pivotDataByYear[fact.Year][dim1Val][dim2Val] = *fact.Value
		}
	}

	// Create a sheet for each year
	firstSheet := true
	for _, year := range years {
		sheetName := sanitizeSheetName(fmt.Sprintf("%d", year))

		if firstSheet {
			if err := f.SetSheetName("Sheet1", sheetName); err != nil {
				log.Printf("failed to set sheet name for year %d: %v", year, err)
				sheetName = "Sheet1"
			}
			firstSheet = false
		} else {
			_, err := f.NewSheet(sheetName)
			if err != nil {
				log.Printf("failed to create sheet for year %d: %v", year, err)
				continue
			}
		}

		currentRow := 1
		pivotData := pivotDataByYear[year]

		// Header row 1: dim1 name (A1:A2) and dim2 name (B1:last column)
		cellA1, _ := excelize.CoordinatesToCellName(1, currentRow)
		cellA2, _ := excelize.CoordinatesToCellName(1, currentRow+1)
		f.SetCellValue(sheetName, cellA1, dim1.Name)
		f.MergeCell(sheetName, cellA1, cellA2)
		f.SetCellStyle(sheetName, cellA1, cellA2, headerStyle)

		cellB1, _ := excelize.CoordinatesToCellName(2, currentRow)
		cellLast1, _ := excelize.CoordinatesToCellName(len(dim2Values)+1, currentRow)
		f.SetCellValue(sheetName, cellB1, dim2.Name)
		f.MergeCell(sheetName, cellB1, cellLast1)
		f.SetCellStyle(sheetName, cellB1, cellLast1, headerStyle)
		currentRow++

		// Header row 2: dim2 values
		for colIdx, dim2Val := range dim2Values {
			cellName, _ := excelize.CoordinatesToCellName(colIdx+2, currentRow)
			f.SetCellValue(sheetName, cellName, dim2Val)
			f.SetCellStyle(sheetName, cellName, cellName, headerStyle)
		}
		currentRow++

		// Data rows
		for _, dim1Val := range dim1Values {
			cellName, _ := excelize.CoordinatesToCellName(1, currentRow)
			f.SetCellValue(sheetName, cellName, dim1Val)
			f.SetCellStyle(sheetName, cellName, cellName, headerStyle)

			for colIdx, dim2Val := range dim2Values {
				cellName, _ := excelize.CoordinatesToCellName(colIdx+2, currentRow)
				if pivotData != nil && pivotData[dim1Val] != nil && pivotData[dim1Val][dim2Val] != nil {
					f.SetCellValue(sheetName, cellName, pivotData[dim1Val][dim2Val])
				} else {
					f.SetCellValue(sheetName, cellName, "")
				}
				f.SetCellStyle(sheetName, cellName, cellName, dataStyle)
			}
			currentRow++
		}

		// Auto-fit columns
		f.SetColWidth(sheetName, "A", "A", 25)
		for i := 0; i < len(dim2Values); i++ {
			colName, _ := excelize.ColumnNumberToName(i + 2)
			f.SetColWidth(sheetName, colName, colName, 15)
		}
	}

	return nil
}

// exportDimension3Plus exports table with 3+ dimensions (standard table)
func (s *ExcelService) exportDimension3Plus(f *excelize.File, table *dto.TableResponse, years []int, headerStyle, dataStyle int) error {
	sheetName := "Sheet1"
	currentRow := 1

	// Build headers
	headers := []string{"No"}
	for _, dim := range table.Dimensions {
		headers = append(headers, dim.Name)
	}
	headers = append(headers, "Year", "Value")

	// Write headers
	for colIdx, header := range headers {
		cellName, _ := excelize.CoordinatesToCellName(colIdx+1, currentRow)
		f.SetCellValue(sheetName, cellName, header)
		f.SetCellStyle(sheetName, cellName, cellName, headerStyle)
	}
	currentRow++

	// Write data rows
	for idx, fact := range table.Facts {
		colIdx := 0

		// Row number
		cellName, _ := excelize.CoordinatesToCellName(colIdx+1, currentRow)
		f.SetCellValue(sheetName, cellName, idx+1)
		f.SetCellStyle(sheetName, cellName, cellName, dataStyle)
		colIdx++

		// Dimension values
		for _, dim := range table.Dimensions {
			var dimValue string
			for _, fv := range fact.Dimensions {
				if fv.ID == dim.ID {
					dimValue = fv.Value.Name
					break
				}
			}
			cellName, _ := excelize.CoordinatesToCellName(colIdx+1, currentRow)
			f.SetCellValue(sheetName, cellName, dimValue)
			f.SetCellStyle(sheetName, cellName, cellName, dataStyle)
			colIdx++
		}

		// Year
		cellName, _ = excelize.CoordinatesToCellName(colIdx+1, currentRow)
		f.SetCellValue(sheetName, cellName, fact.Year)
		f.SetCellStyle(sheetName, cellName, cellName, dataStyle)
		colIdx++

		// Value
		cellName, _ = excelize.CoordinatesToCellName(colIdx+1, currentRow)
		if fact.Value != nil {
			f.SetCellValue(sheetName, cellName, *fact.Value)
		} else {
			f.SetCellValue(sheetName, cellName, "")
		}
		f.SetCellStyle(sheetName, cellName, cellName, dataStyle)

		currentRow++
	}

	// Auto-fit columns
	for i := 0; i < len(headers); i++ {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheetName, colName, colName, 15)
	}

	return nil
}

// sanitizeSheetName ensures the sheet name is valid for Excel
func sanitizeSheetName(name string) string {
	// Replace invalid characters
	invalidChars := []string{"\\", "/", "?", "*", "[", "]", ":"}
	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, char, "_")
	}

	// Truncate to 31 characters (Excel limit)
	if len(name) > 31 {
		name = name[:31]
	}

	// Ensure not empty
	if name == "" {
		name = "Sheet"
	}

	return name
}
