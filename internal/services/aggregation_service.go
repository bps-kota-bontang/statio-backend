package services

import (
	"fmt"
	"log"
	"sort"
	"statio/internal/dto"
	"statio/internal/models"
	"statio/internal/repositories"
	"strings"

	"gorm.io/gorm"
)

type AggregationService struct {
	db            *gorm.DB
	tableRepo     repositories.TableRepository
	dimensionRepo repositories.DimensionRepository
	factRepo      repositories.FactRepository
}

func NewAggregationService(
	db *gorm.DB,
	tableRepo repositories.TableRepository,
	dimensionRepo repositories.DimensionRepository,
	factRepo repositories.FactRepository,
) *AggregationService {
	return &AggregationService{
		db:            db,
		tableRepo:     tableRepo,
		dimensionRepo: dimensionRepo,
		factRepo:      factRepo,
	}
}

// GenerateParentTable membuat atau update tabel parent berdasarkan agregasi dari child table
func (s *AggregationService) GenerateParentTable(tableID string, req *dto.GenerateParentTableRequest) (*dto.GenerateParentTableResponse, error) {
	// 1. Validasi dan ambil child table
	childTable, err := s.tableRepo.FindDetailedByID(tableID)
	if err != nil {
		return nil, fmt.Errorf("failed to find child table: %w", err)
	}

	if len(childTable.Dimensions) == 0 {
		return nil, fmt.Errorf("table has no dimensions")
	}

	// 2. Validasi semua dimension IDs ada di table dan load dimensions dengan parent hierarchy
	aggregationInfo, err := s.validateAndLoadDimensions(childTable, req.DimensionIDs)
	if err != nil {
		return nil, err
	}

	if len(aggregationInfo) == 0 {
		return nil, fmt.Errorf("none of the specified dimensions have parent hierarchy")
	}

	log.Printf("Processing %d dimensions for aggregation", len(aggregationInfo))
	for _, info := range aggregationInfo {
		log.Printf("- Dimension: %s (ID: %s) with %d parent values",
			info.Dimension.Name, info.Dimension.ID, len(info.ParentValueIDs))
	}

	// 3. Cari atau buat parent table
	parentTable, isNewTable, err := s.findOrCreateParentTableMultiDim(childTable, aggregationInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to find or create parent table: %w", err)
	}

	// 4. Ambil semua facts dari child table
	childFacts, err := s.factRepo.FindAllByTable(tableID)
	if err != nil {
		return nil, fmt.Errorf("failed to find child facts: %w", err)
	}

	// 5. Aggregate facts untuk parent table dengan multiple dimensions
	err = s.aggregateAndSaveFactsMultiDim(parentTable, childTable, childFacts, aggregationInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate facts: %w", err)
	}

	// Build response
	aggregatedDimensions := make([]dto.AggregatedDimensionInfo, len(aggregationInfo))
	for i, info := range aggregationInfo {
		aggregatedDimensions[i] = dto.AggregatedDimensionInfo{
			DimensionID:      info.Dimension.ID,
			DimensionName:    info.Dimension.Name,
			ParentValuesUsed: len(info.ParentValueIDs),
		}
	}

	return &dto.GenerateParentTableResponse{
		ParentTableID:        parentTable.ID,
		IsNewTable:           isNewTable,
		Message:              s.buildResponseMessage(isNewTable, parentTable.Name),
		ChildTableID:         tableID,
		AggregatedDimensions: aggregatedDimensions,
	}, nil
}

// DimensionAggregationInfo menyimpan informasi untuk agregasi satu dimension
type DimensionAggregationInfo struct {
	Dimension      *models.Dimension
	ParentValueIDs []string
	ParentAggMap   map[string][]string // parent_id -> []child_id
}

// validateAndLoadDimensions validasi dan load dimensions yang akan diagregasi
func (s *AggregationService) validateAndLoadDimensions(
	childTable *models.Table,
	dimensionIDs []string,
) ([]*DimensionAggregationInfo, error) {
	// Build map table dimensions untuk cepat lookup
	tableDimMap := make(map[string]bool)
	for _, td := range childTable.Dimensions {
		tableDimMap[td.DimensionID] = true
	}

	var aggregationInfo []*DimensionAggregationInfo

	for _, dimID := range dimensionIDs {
		// Validasi dimension ada di table
		if !tableDimMap[dimID] {
			return nil, fmt.Errorf("dimension %s not found in table", dimID)
		}

		// Load dimension dengan values
		dimension, err := s.dimensionRepo.FindByIDWithValues(dimID)
		if err != nil {
			return nil, fmt.Errorf("failed to load dimension %s: %w", dimID, err)
		}

		// Check apakah dimension ini memiliki parent hierarchy
		hasParentHierarchy := false
		for _, val := range dimension.Values {
			if val.ParentID != nil {
				hasParentHierarchy = true
				break
			}
		}

		if !hasParentHierarchy {
			log.Printf("Warning: dimension %s (%s) has no parent hierarchy, skipping", dimension.Name, dimID)
			continue
		}

		// Ambil parent value IDs
		parentValueIDs, err := s.dimensionRepo.FindParentValueIDs(dimID)
		if err != nil {
			return nil, fmt.Errorf("failed to find parent value IDs for dimension %s: %w", dimID, err)
		}

		if len(parentValueIDs) == 0 {
			log.Printf("Warning: dimension %s (%s) has no parent values, skipping", dimension.Name, dimID)
			continue
		}

		// Build parent aggregation map
		parentAggMap := s.buildParentAggregationMap(dimension.Values)

		aggregationInfo = append(aggregationInfo, &DimensionAggregationInfo{
			Dimension:      dimension,
			ParentValueIDs: parentValueIDs,
			ParentAggMap:   parentAggMap,
		})
	}

	return aggregationInfo, nil
}

// buildParentAggregationMap membuat map parent_id -> []child_id
func (s *AggregationService) buildParentAggregationMap(values []models.DimensionValue) map[string][]string {
	parentAggMap := make(map[string][]string)

	for _, val := range values {
		if val.ParentID != nil {
			parentID := *val.ParentID
			if _, exists := parentAggMap[parentID]; !exists {
				parentAggMap[parentID] = []string{}
			}
			parentAggMap[parentID] = append(parentAggMap[parentID], val.ID)
		}
	}

	return parentAggMap
}

// findOrCreateParentTableMultiDim mencari parent table yang sudah ada atau membuat yang baru (untuk multiple dimensions)
func (s *AggregationService) findOrCreateParentTableMultiDim(
	childTable *models.Table,
	aggregationInfo []*DimensionAggregationInfo,
) (*models.Table, bool, error) {
	// Generate parent table name berdasarkan semua dimensions yang diagregasi
	// Collect child -> parent dimension name mappings
	dimNameReplacements := make(map[string]string) // child_name -> parent_name
	aggregatedDimIDs := make(map[string]bool)
	for _, info := range aggregationInfo {
		childDimName := info.Dimension.Name
		aggregatedDimIDs[info.Dimension.ID] = true

		// Check if parent values use different dimension
		parentDimID := s.getParentDimensionID(info.Dimension, info.ParentValueIDs)
		if parentDimID != "" {
			// Load parent dimension to get its name
			parentDim, err := s.dimensionRepo.FindDimensionByID(parentDimID)
			if err != nil {
				log.Printf("Warning: failed to load parent dimension: %v", err)
				dimNameReplacements[childDimName] = childDimName // fallback to child name
			} else {
				dimNameReplacements[childDimName] = parentDim.Name
			}
		} else {
			// Same dimension, use child name (will be replaced with parent level values)
			dimNameReplacements[childDimName] = childDimName
		}
	}

	parentTableName := s.generateParentTableNameMultiDim(childTable.Name, dimNameReplacements)

	// Build parent dimensions - gunakan parent dimension jika ada
	var parentDimensions []models.TableDimension
	for i, td := range childTable.Dimensions {
		// Check apakah dimension ini diagregasi
		dimensionID := td.DimensionID

		// Cari apakah ada di aggregationInfo
		for _, info := range aggregationInfo {
			if info.Dimension.ID == td.DimensionID {
				// Dimension ini diagregasi, cek apakah parent values punya dimension berbeda
				parentDimID := s.getParentDimensionID(info.Dimension, info.ParentValueIDs)
				if parentDimID != "" {
					dimensionID = parentDimID
				}
				break
			}
		}

		parentDimensions = append(parentDimensions, models.TableDimension{
			DimensionID: dimensionID,
			Order:       i,
		})
	}

	// Check apakah parent table sudah ada dengan source_table_id yang sama
	existingParentTables, err := s.tableRepo.FindAllBySourceTableID(childTable.ID)

	if err == nil && len(existingParentTables) > 0 {
		// Cari table yang dimensions-nya match
		for _, existingTable := range existingParentTables {
			if s.isDimensionsMatch(existingTable.Dimensions, parentDimensions) {
				log.Printf("Found existing parent table: %s (ID: %s), will update facts only", existingTable.Name, existingTable.ID)

				// Delete old facts
				if err := s.deleteTableFacts(existingTable); err != nil {
					return nil, false, fmt.Errorf("failed to delete old facts: %w", err)
				}

				return existingTable, false, nil
			}
		}
		log.Printf("Found %d parent table(s) but none with matching dimensions, creating new table", len(existingParentTables))
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, false, fmt.Errorf("failed to check existing parent table: %w", err)
	}

	// Simpan child table ID di Notes untuk traceability
	childTableIDNote := fmt.Sprintf("Generated from child table ID: %s", childTable.ID)
	if childTable.Notes != nil && *childTable.Notes != "" {
		combined := fmt.Sprintf("%s\n\n%s", childTableIDNote, *childTable.Notes)
		childTableIDNote = combined
	}

	// Buat table baru
	parentTable := &models.Table{
		Name:           parentTableName,
		Direction:      childTable.Direction,
		Description:    s.generateParentDescriptionMultiDim(childTable, dimNameReplacements),
		IndicatorID:    childTable.IndicatorID,
		OrganizationID: childTable.OrganizationID,
		Labels:         childTable.Labels,
		Notes:          &childTableIDNote,
		Status:         "draft",
		IsLocked:       false,
		IsAggregated:   true,
		SourceTableID:  &childTable.ID,
	}

	// Mulai transaction
	tx := s.tableRepo.BeginTx()
	if tx.Error != nil {
		return nil, false, tx.Error
	}

	// Simpan parent table
	if err := s.tableRepo.CreateWithTx(tx, parentTable); err != nil {
		tx.Rollback()
		return nil, false, fmt.Errorf("failed to create parent table: %w", err)
	}

	// Simpan table dimensions
	for i := range parentDimensions {
		parentDimensions[i].TableID = parentTable.ID
		if err := s.tableRepo.CreateTableDimensionWithTx(tx, &parentDimensions[i]); err != nil {
			tx.Rollback()
			return nil, false, fmt.Errorf("failed to create table dimension: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, false, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Created parent table: %s (ID: %s)", parentTable.Name, parentTable.ID)

	return parentTable, true, nil
}

// findOrCreateParentTable mencari parent table yang sudah ada atau membuat yang baru
func (s *AggregationService) findOrCreateParentTable(
	childTable *models.Table,
	dimension *models.Dimension,
	parentValueIDs []string,
) (*models.Table, bool, error) {
	// Generate parent table name
	parentTableName := s.generateParentTableName(childTable.Name, dimension.Name)

	// Cari tabel dengan nama yang sama dan indicator yang sama
	// Untuk simplifikasi, kita buat table baru dengan naming convention
	// TODO: Implementasi logika untuk cek apakah parent table sudah ada

	// Untuk saat ini, kita buat table baru
	parentTable := &models.Table{
		Name:           parentTableName,
		Direction:      childTable.Direction,
		Description:    s.generateParentDescription(childTable),
		IndicatorID:    childTable.IndicatorID,
		OrganizationID: childTable.OrganizationID,
		Labels:         childTable.Labels,
		Status:         "draft",
		IsLocked:       false,
	}

	// Buat dimensions untuk parent table (copy dari child, tapi ganti target dimension dengan parent values)
	var parentDimensions []models.TableDimension
	for i, td := range childTable.Dimensions {
		if td.DimensionID == dimension.ID {
			// Untuk dimension yang diagregasi, tetap gunakan dimension yang sama
			// Karena parent values masih bagian dari dimension yang sama
			parentDimensions = append(parentDimensions, models.TableDimension{
				DimensionID: td.DimensionID,
				Order:       i,
			})
		} else {
			// Dimensi lain tetap sama
			parentDimensions = append(parentDimensions, models.TableDimension{
				DimensionID: td.DimensionID,
				Order:       i,
			})
		}
	}

	// Mulai transaction
	tx := s.tableRepo.BeginTx()
	if tx.Error != nil {
		return nil, false, tx.Error
	}

	// Simpan parent table
	if err := s.tableRepo.CreateWithTx(tx, parentTable); err != nil {
		tx.Rollback()
		return nil, false, fmt.Errorf("failed to create parent table: %w", err)
	}

	// Simpan table dimensions
	for i := range parentDimensions {
		parentDimensions[i].TableID = parentTable.ID
		if err := s.tableRepo.CreateTableDimensionWithTx(tx, &parentDimensions[i]); err != nil {
			tx.Rollback()
			return nil, false, fmt.Errorf("failed to create table dimension: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, false, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Created parent table: %s (ID: %s)", parentTable.Name, parentTable.ID)

	return parentTable, true, nil
}

// aggregateAndSaveFactsMultiDim melakukan agregasi facts dari child ke parent (multiple dimensions)
func (s *AggregationService) aggregateAndSaveFactsMultiDim(
	parentTable *models.Table,
	childTable *models.Table,
	childFacts []*models.Fact,
	aggregationInfo []*DimensionAggregationInfo,
) error {
	// Build map dimension_id -> aggregation info untuk cepat lookup
	aggInfoMap := make(map[string]*DimensionAggregationInfo)
	childToParentMaps := make(map[string]map[string]string) // dimension_id -> (child_id -> parent_id)

	for _, info := range aggregationInfo {
		aggInfoMap[info.Dimension.ID] = info

		// Build reverse map: child_id -> parent_id untuk dimension ini
		childToParentMap := make(map[string]string)
		for parentID, childIDs := range info.ParentAggMap {
			for _, childID := range childIDs {
				childToParentMap[childID] = parentID
			}
		}
		childToParentMaps[info.Dimension.ID] = childToParentMap
	}

	// Group facts by: year + parent_values_for_all_aggregated_dims + other_dimension_values
	type factKey struct {
		year       int
		parentDims string // serialized parent dimension value IDs
		otherDims  string // serialized other dimension value IDs
	}

	factGroups := make(map[factKey][]*models.Fact)

	// Load all fact dimension values for child facts
	childFactIDs := make([]string, len(childFacts))
	for i, f := range childFacts {
		childFactIDs[i] = f.ID
	}

	// Map fact_id -> []FactDimensionValue
	factDimValuesMap := make(map[string][]models.FactDimensionValue)
	allFactDimValues, err := s.factRepo.FindFactDimensionValuesByFactIDs(childFactIDs)
	if err != nil {
		return fmt.Errorf("failed to load fact dimension values: %w", err)
	}

	for _, fdv := range allFactDimValues {
		factDimValuesMap[fdv.FactID] = append(factDimValuesMap[fdv.FactID], fdv)
	}

	// Group facts
	for _, fact := range childFacts {
		factDimValues := factDimValuesMap[fact.ID]

		parentDimMap := make(map[string]string) // dimension_id -> parent_value_id
		otherDimIDs := []string{}
		skipFact := false

		for _, fdv := range factDimValues {
			if fdv.DimensionValue == nil {
				continue
			}

			dimID := fdv.DimensionValue.DimensionID

			// Check apakah dimension ini perlu diagregasi
			if childToParentMap, isAggregated := childToParentMaps[dimID]; isAggregated {
				// Dimension yang diagregasi - cari parent-nya
				if parentID, exists := childToParentMap[fdv.DimensionValueID]; exists {
					parentDimMap[dimID] = parentID
				} else {
					// Jika tidak ada parent, skip fact ini
					log.Printf("Warning: fact %s has dimension value %s without parent in dimension %s",
						fact.ID, fdv.DimensionValueID, dimID)
					skipFact = true
					break
				}
			} else {
				// Dimensi lain yang tidak diagregasi
				otherDimIDs = append(otherDimIDs, fdv.DimensionValueID)
			}
		}

		if skipFact || len(parentDimMap) != len(aggregationInfo) {
			continue
		}

		// Serialize parent dimension values (sorted by dimension ID untuk consistency)
		parentDimSlice := make([]string, 0, len(parentDimMap))
		for dimID, parentValID := range parentDimMap {
			parentDimSlice = append(parentDimSlice, fmt.Sprintf("%s:%s", dimID, parentValID))
		}
		sort.Strings(parentDimSlice)
		sort.Strings(otherDimIDs)

		key := factKey{
			year:       fact.Year,
			parentDims: strings.Join(parentDimSlice, "|"),
			otherDims:  strings.Join(otherDimIDs, "|"),
		}

		factGroups[key] = append(factGroups[key], fact)
	}

	// Aggregate dan simpan facts
	tx := s.factRepo.BeginTx()
	if tx.Error != nil {
		return tx.Error
	}

	for key, facts := range factGroups {
		// Aggregate values
		var totalValue float64
		var totalOldValue float64
		hasOldValue := false

		for _, f := range facts {
			if f.Value != nil {
				totalValue += *f.Value
			}
			if f.OldValue != nil {
				totalOldValue += *f.OldValue
				hasOldValue = true
			}
		}

		// Create parent fact
		parentFact := models.Fact{
			TableID: parentTable.ID,
			Year:    key.year,
			Value:   &totalValue,
		}

		if hasOldValue {
			parentFact.OldValue = &totalOldValue
		}

		if err := s.factRepo.CreateFactWithTx(tx, &parentFact); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create parent fact: %w", err)
		}

		// Create fact dimension values
		// 1. Parent dimension values
		parentDimPairs := strings.Split(key.parentDims, "|")
		for _, pair := range parentDimPairs {
			parts := strings.Split(pair, ":")
			if len(parts) == 2 {
				parentValID := parts[1]
				parentFactDimVal := models.FactDimensionValue{
					FactID:           parentFact.ID,
					DimensionValueID: parentValID,
				}
				if err := s.factRepo.CreateFactDimensionValueWithTx(tx, &parentFactDimVal); err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to create parent fact dimension value: %w", err)
				}
			}
		}

		// 2. Other dimension values (yang tidak diagregasi)
		if key.otherDims != "" {
			otherDimIDs := strings.Split(key.otherDims, "|")
			for _, dimValID := range otherDimIDs {
				otherFactDimVal := models.FactDimensionValue{
					FactID:           parentFact.ID,
					DimensionValueID: dimValID,
				}
				if err := s.factRepo.CreateFactDimensionValueWithTx(tx, &otherFactDimVal); err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to create fact dimension value: %w", err)
				}
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit facts: %w", err)
	}

	log.Printf("Aggregated %d fact groups into parent table", len(factGroups))

	return nil
}

// aggregateAndSaveFacts melakukan agregasi facts dari child ke parent
func (s *AggregationService) aggregateAndSaveFacts(
	parentTable *models.Table,
	childTable *models.Table,
	childFacts []*models.Fact,
	targetDimensionID string,
	parentAggMap map[string][]string,
) error {
	// Group facts by: year + other dimensions + parent_value
	type factKey struct {
		year      int
		parentID  string
		otherDims string // serialized other dimension IDs
	}

	factGroups := make(map[factKey][]*models.Fact)

	// Load all fact dimension values for child facts
	childFactIDs := make([]string, len(childFacts))
	for i, f := range childFacts {
		childFactIDs[i] = f.ID
	}

	// Map fact_id -> []FactDimensionValue
	factDimValuesMap := make(map[string][]models.FactDimensionValue)
	allFactDimValues, err := s.factRepo.FindFactDimensionValuesByFactIDs(childFactIDs)
	if err != nil {
		return fmt.Errorf("failed to load fact dimension values: %w", err)
	}

	for _, fdv := range allFactDimValues {
		factDimValuesMap[fdv.FactID] = append(factDimValuesMap[fdv.FactID], fdv)
	}

	// Build reverse map: child_id -> parent_id
	childToParentMap := make(map[string]string)
	for parentID, childIDs := range parentAggMap {
		for _, childID := range childIDs {
			childToParentMap[childID] = parentID
		}
	}

	// Group facts
	for _, fact := range childFacts {
		factDimValues := factDimValuesMap[fact.ID]

		var parentID string
		otherDimIDs := []string{}

		for _, fdv := range factDimValues {
			if fdv.DimensionValue == nil {
				continue
			}

			if fdv.DimensionValue.DimensionID == targetDimensionID {
				// Ini dimension yang diagregasi - cari parent-nya
				if pID, exists := childToParentMap[fdv.DimensionValueID]; exists {
					parentID = pID
				} else {
					// Jika tidak ada parent, skip fact ini
					log.Printf("Warning: fact %s has dimension value %s without parent", fact.ID, fdv.DimensionValueID)
					continue
				}
			} else {
				// Dimensi lain
				otherDimIDs = append(otherDimIDs, fdv.DimensionValueID)
			}
		}

		if parentID == "" {
			continue
		}

		key := factKey{
			year:      fact.Year,
			parentID:  parentID,
			otherDims: strings.Join(otherDimIDs, "|"),
		}

		factGroups[key] = append(factGroups[key], fact)
	}

	// Aggregate dan simpan facts
	tx := s.factRepo.BeginTx()
	if tx.Error != nil {
		return tx.Error
	}

	for key, facts := range factGroups {
		// Aggregate values
		var totalValue float64
		var totalOldValue float64
		hasOldValue := false

		for _, f := range facts {
			if f.Value != nil {
				totalValue += *f.Value
			}
			if f.OldValue != nil {
				totalOldValue += *f.OldValue
				hasOldValue = true
			}
		}

		// Create parent fact
		parentFact := models.Fact{
			TableID: parentTable.ID,
			Year:    key.year,
			Value:   &totalValue,
		}

		if hasOldValue {
			parentFact.OldValue = &totalOldValue
		}

		if err := s.factRepo.CreateFactWithTx(tx, &parentFact); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create parent fact: %w", err)
		}

		// Create fact dimension values
		// 1. Parent dimension value
		parentFactDimVal := models.FactDimensionValue{
			FactID:           parentFact.ID,
			DimensionValueID: key.parentID,
		}
		if err := s.factRepo.CreateFactDimensionValueWithTx(tx, &parentFactDimVal); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create parent fact dimension value: %w", err)
		}

		// 2. Other dimension values
		if key.otherDims != "" {
			otherDimIDs := strings.Split(key.otherDims, "|")
			for _, dimValID := range otherDimIDs {
				otherFactDimVal := models.FactDimensionValue{
					FactID:           parentFact.ID,
					DimensionValueID: dimValID,
				}
				if err := s.factRepo.CreateFactDimensionValueWithTx(tx, &otherFactDimVal); err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to create fact dimension value: %w", err)
				}
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit facts: %w", err)
	}

	log.Printf("Aggregated %d fact groups into parent table", len(factGroups))

	return nil
}

// generateParentTableName membuat nama untuk parent table
func (s *AggregationService) generateParentTableName(childTableName, dimensionName string) string {
	// Contoh: "Jumlah Penduduk Menurut Kelurahan" -> "Jumlah Penduduk Menurut Kecamatan"
	// Untuk simplifikasi, tambahkan prefix
	return fmt.Sprintf("%s (Agregasi %s - Parent Level)", childTableName, dimensionName)
}

// generateParentTableNameMultiDim membuat nama untuk parent table (multiple dimensions)
func (s *AggregationService) generateParentTableNameMultiDim(childTableName string, dimNameReplacements map[string]string) string {
	// Contoh: "Jumlah Penduduk Menurut Kelurahan dan Agama" ->
	//         "Jumlah Penduduk Menurut Kecamatan dan Agama"
	// Replace child dimension names with parent dimension names
	result := childTableName
	for childName, parentName := range dimNameReplacements {
		// Case-insensitive replacement
		result = strings.ReplaceAll(result, childName, parentName)
	}
	return result
}

// generateParentDescription membuat deskripsi untuk parent table
func (s *AggregationService) generateParentDescription(childTable *models.Table) *string {
	desc := fmt.Sprintf("Tabel agregasi parent dari tabel: %s", childTable.Name)
	if childTable.Description != nil {
		desc += fmt.Sprintf(". Deskripsi child table: %s", *childTable.Description)
	}
	return &desc
}

// generateParentDescriptionMultiDim membuat deskripsi untuk parent table (multiple dimensions)
func (s *AggregationService) generateParentDescriptionMultiDim(childTable *models.Table, dimNameReplacements map[string]string) *string {
	// Build dimension replacement info
	replacements := []string{}
	for childName, parentName := range dimNameReplacements {
		if childName != parentName {
			replacements = append(replacements, fmt.Sprintf("%s → %s", childName, parentName))
		}
	}

	desc := fmt.Sprintf("Tabel agregasi parent dari tabel: %s", childTable.Name)
	if len(replacements) > 0 {
		desc += fmt.Sprintf(". Dimensi yang diagregasi: %s", strings.Join(replacements, ", "))
	}
	if childTable.Description != nil {
		desc += fmt.Sprintf(". Deskripsi child table: %s", *childTable.Description)
	}
	return &desc
}

// buildResponseMessage membuat pesan response
func (s *AggregationService) buildResponseMessage(isNewTable bool, tableName string) string {
	if isNewTable {
		return fmt.Sprintf("Parent table '%s' berhasil dibuat dengan data agregasi", tableName)
	}
	return fmt.Sprintf("Parent table '%s' berhasil diupdate dengan data agregasi terbaru", tableName)
}

// isDimensionsMatch memeriksa apakah dimensions dari dua table match
func (s *AggregationService) isDimensionsMatch(existingDims []models.TableDimension, newDims []models.TableDimension) bool {
	if len(existingDims) != len(newDims) {
		return false
	}

	// Sort by order untuk memastikan perbandingan yang benar
	existingMap := make(map[int]string)
	newMap := make(map[int]string)

	for _, d := range existingDims {
		existingMap[d.Order] = d.DimensionID
	}

	for _, d := range newDims {
		newMap[d.Order] = d.DimensionID
	}

	for order, dimID := range existingMap {
		if newMap[order] != dimID {
			return false
		}
	}

	return true
}

// deleteTableFacts menghapus semua facts dari table
func (s *AggregationService) deleteTableFacts(table *models.Table) error {
	// Get all facts
	facts, err := s.factRepo.FindFactsByTableID(table.ID)
	if err != nil {
		return fmt.Errorf("failed to find facts: %w", err)
	}

	if len(facts) == 0 {
		return nil
	}

	factIDs := make([]string, len(facts))
	for i, f := range facts {
		factIDs[i] = f.ID
	}

	// Delete fact dimension values first
	if err := s.factRepo.DeleteFactDimensionValuesByFactIDs(factIDs); err != nil {
		return fmt.Errorf("failed to delete fact dimension values: %w", err)
	}

	// Delete facts
	if err := s.factRepo.DeleteFactsByIDs(factIDs); err != nil {
		return fmt.Errorf("failed to delete facts: %w", err)
	}

	log.Printf("Deleted %d facts from table %s", len(facts), table.ID)
	return nil
}

// getParentDimensionID mendapatkan parent dimension ID jika parent values berada di dimension berbeda
func (s *AggregationService) getParentDimensionID(dimension *models.Dimension, parentValueIDs []string) string {
	if len(parentValueIDs) == 0 {
		return ""
	}

	// Ambil parent dimension value pertama untuk cek
	parentValue, err := s.factRepo.FindDimensionValueByID(parentValueIDs[0])
	if err != nil {
		log.Printf("Warning: failed to load parent dimension value: %v", err)
		return ""
	}

	// Jika parent value memiliki dimension ID berbeda dengan child dimension, return parent dimension ID
	if parentValue.DimensionID != dimension.ID {
		log.Printf("Parent values use different dimension: %s (child: %s)", parentValue.DimensionID, dimension.ID)
		return parentValue.DimensionID
	}

	return ""
}
