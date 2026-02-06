package handlers

import (
	"log"
	"statio/internal/dto"
	"statio/internal/services"
	"statio/utils"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type TableHandler struct {
	service  *services.TableService
	validate *validator.Validate
}

func NewTableHandler(service *services.TableService, validate *validator.Validate) *TableHandler {
	return &TableHandler{
		service:  service,
		validate: validate,
	}
}

func (h *TableHandler) GetAllTables(c *fiber.Ctx) error {
	sortBy := c.Query("sort_by", "no")
	sortOrder := c.Query("sort_order", "asc")
	search := c.Query("search")
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 10)
	roles := c.Locals("roles").([]string)
	organizationID := c.Locals("organization_id").(*string)

	// Ambil filters per kolom, bisa multiple
	filters := map[string][]string{}
	keys := []string{
		"indicator_name",
		"indicator_measure",
		"indicator_unit",
		"dimensions",
		"organization_id",
		"labels",
		"missing_facts",
		"outlier_facts",
		"revision_facts",
		"status",
		"is_aggregated",
		"is_show",
		"is_integrated",
		"direction",
	}
	for _, key := range keys {
		// c.Context().QueryArgs().PeekMulti(key) mengembalikan [][]byte
		values := c.Context().QueryArgs().PeekMulti(key)
		if len(values) > 0 {
			strs := make([]string, len(values))
			for i, v := range values {
				strs[i] = string(v)
			}
			filters[key] = strs
		}
	}

	if utils.IsOperator(roles) {
		if organizationID != nil {
			filters["organization_id"] = []string{*organizationID}
			filters["is_aggregated"] = []string{"false"}
			filters["is_show"] = []string{"true"}
		} else {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"data":    nil,
				"message": "Missing organization context for operator access",
			})
		}
	}

	tables, total, err := h.service.GetAllPaginated(search, page, perPage, sortBy, sortOrder, filters)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	meta := utils.NewPaginationMeta(total, page, perPage)

	return c.JSON(fiber.Map{
		"data":    tables,
		"message": "Tables fetched successfully",
		"meta":    meta,
	})
}

func (h *TableHandler) GetTable(c *fiber.Ctx) error {
	id := c.Params("id")
	roles := c.Locals("roles").([]string)
	orgID := c.Locals("organization_id").(*string)
	yearParam := c.QueryInt("year")

	var year *int
	if yearParam != 0 {
		year = &yearParam
	}

	table, err := h.service.GetByID(id, year)
	if err != nil {
		status := 500
		if err == gorm.ErrRecordNotFound {
			status = 404
		}
		return c.Status(status).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	if utils.IsOperator(roles) {
		if orgID == nil || table.Organization == nil || table.Organization.ID != *orgID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"data":    nil,
				"message": "You are not authorized to access this table",
			})
		}
	}

	return c.JSON(fiber.Map{
		"data":    table,
		"message": "Table fetched successfully",
	})
}

func (h *TableHandler) UpdateFacts(c *fiber.Ctx) error {
	id := c.Params("id")
	roles := c.Locals("roles").([]string)
	orgID := c.Locals("organization_id").(*string)

	var payload dto.UpdateFactRequest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid request payload",
		})
	}

	if err := h.validate.Struct(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	if err := h.service.UpdateTableFacts(id, &payload, roles, orgID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    nil,
		"message": "Facts updated successfully",
	})
}

func (h *TableHandler) GetFacts(c *fiber.Ctx) error {
	id := c.Params("id")
	roles := c.Locals("roles").([]string)
	orgID := c.Locals("organization_id").(*string)

	dimValueIDs := c.Context().QueryArgs().PeekMulti("dimension_value_ids")
	var dims []string
	for _, d := range dimValueIDs {
		dims = append(dims, string(d))
	}

	facts, err := h.service.GetTableFacts(id, dims, roles, orgID)
	if err != nil {
		status := 500
		if err == gorm.ErrRecordNotFound {
			status = 404
		}
		return c.Status(status).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    facts,
		"message": "Facts fetched successfully",
	})
}

func (h *TableHandler) CreateTable(c *fiber.Ctx) error {
	var payload dto.CreateTableRequest
	roles := c.Locals("roles").([]string)

	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to create table",
		})
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid request payload",
		})
	}

	if err := h.validate.Struct(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	table, err := h.service.Create(&payload)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"data":    table,
		"message": "Table created successfully",
	})
}

func (h *TableHandler) UpdateTable(c *fiber.Ctx) error {
	id := c.Params("id")
	roles := c.Locals("roles").([]string)

	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to update table",
		})
	}

	var payload dto.UpdateTableRequest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid request body",
		})
	}

	if err := h.validate.Struct(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	if err := h.service.UpdateWithRelations(id, &payload); err != nil {
		status := 500
		if err == gorm.ErrRecordNotFound {
			status = 404
		}
		return c.Status(status).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    nil,
		"message": "Table updated successfully",
	})
}

func (h *TableHandler) GetTableLabels(c *fiber.Ctx) error {
	response, err := h.service.GetAllLabels()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    response,
		"message": "Table labels fetched successfully",
	})
}

func (h *TableHandler) AddLabelsToTables(c *fiber.Ctx) error {
	roles := c.Locals("roles").([]string)
	orgID := c.Locals("organization_id").(*string)

	var payload dto.AddLabelsToTablesRequest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid request payload",
		})
	}

	if err := h.validate.Struct(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	if err := h.service.AddLabelsBulk(&payload, roles, orgID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    nil,
		"message": "Labels added to tables successfully",
	})
}

func (h *TableHandler) UpdateLabels(c *fiber.Ctx) error {
	roles := c.Locals("roles").([]string)
	orgID := c.Locals("organization_id").(*string)

	id := c.Params("id")

	var payload dto.UpdateTableLabelsRequest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid request payload",
		})
	}

	if err := h.validate.Struct(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	if err := h.service.UpdateTableLabels(id, &payload, roles, orgID); err != nil {
		status := 500
		if err == gorm.ErrRecordNotFound {
			status = 404
		}
		return c.Status(status).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    nil,
		"message": "Labels updated successfully",
	})
}

func (h *TableHandler) UpdateTableName(c *fiber.Ctx) error {
	id := c.Params("id")
	roles := c.Locals("roles").([]string)
	orgID := c.Locals("organization_id").(*string)

	var payload dto.UpdateTableNameRequest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid request payload",
		})
	}

	if err := h.validate.Struct(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	if err := h.service.UpdateTableName(id, payload.Name, roles, orgID); err != nil {
		status := 500
		if err == gorm.ErrRecordNotFound {
			status = 404
		}
		return c.Status(status).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    nil,
		"message": "Table name updated successfully",
	})
}

func (h *TableHandler) UpdateTableNotes(c *fiber.Ctx) error {
	id := c.Params("id")
	roles := c.Locals("roles").([]string)
	orgID := c.Locals("organization_id").(*string)

	var payload dto.UpdateTableNotesRequest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid request payload",
		})
	}

	if err := h.validate.Struct(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	if err := h.service.UpdateTableNotes(id, payload.Notes, roles, orgID); err != nil {
		status := 500
		if err == gorm.ErrRecordNotFound {
			status = 404
		}
		return c.Status(status).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    nil,
		"message": "Table notes updated successfully",
	})
}

func (h *TableHandler) UpdateTableIsLocked(c *fiber.Ctx) error {
	id := c.Params("id")
	roles := c.Locals("roles").([]string)

	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to update table lock status",
		})
	}

	var payload dto.UpdateTableIsLockedRequest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid request payload",
		})
	}

	if err := h.validate.Struct(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	if err := h.service.UpdateTableIsLocked(id, payload.Locked); err != nil {
		status := 500
		if err == gorm.ErrRecordNotFound {
			status = 404
		}
		return c.Status(status).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    nil,
		"message": "Table is_locked status updated successfully",
	})
}

func (h *TableHandler) UpdateTableStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	roles := c.Locals("roles").([]string)
	orgID := c.Locals("organization_id").(*string)

	var payload dto.UpdateTableStatusRequest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid request payload",
		})
	}

	if err := h.validate.Struct(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	if err := h.service.UpdateTableStatus(id, payload.Status, roles, orgID); err != nil {
		status := 500
		if err == gorm.ErrRecordNotFound {
			status = 404
		}
		return c.Status(status).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    nil,
		"message": "Table status updated successfully",
	})
}

func (h *TableHandler) UpdateTableMapping(c *fiber.Ctx) error {
	id := c.Params("id")
	roles := c.Locals("roles").([]string)

	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to update table mapping",
		})
	}

	var payload dto.UpdateTableMappingRequest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid request payload",
		})
	}

	if err := h.validate.Struct(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	if err := h.service.UpdateTableMapping(id, &payload); err != nil {
		status := 500
		if err == gorm.ErrRecordNotFound {
			status = 404
		}
		return c.Status(status).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    nil,
		"message": "Table mapping updated successfully",
	})
}

func (h *TableHandler) GetInsightFacts(c *fiber.Ctx) error {
	id := c.Params("id")
	roles := c.Locals("roles").([]string)
	orgID := c.Locals("organization_id").(*string)
	currentYear := time.Now().Year()
	fromYear := c.QueryInt("from_year", currentYear-4)
	toYear := c.QueryInt("to_year", currentYear)

	insightFacts, err := h.service.GetInsightFacts(id, roles, orgID, fromYear, toYear)
	if err != nil {
		status := 500
		if err == gorm.ErrRecordNotFound {
			status = 404
		}
		return c.Status(status).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    insightFacts,
		"message": "Insight facts fetched successfully",
	})
}

func (h *TableHandler) AnalyzeTable(c *fiber.Ctx) error {
	id := c.Params("id")
	roles := c.Locals("roles").([]string)
	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to analyze table",
		})
	}

	if err := h.service.AnalyzeTable(id); err != nil {
		status := 500
		if err == gorm.ErrRecordNotFound {
			status = 404
		}
		return c.Status(status).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    nil,
		"message": "Table analysis started successfully",
	})
}

func (h *TableHandler) AnalyzeTables(c *fiber.Ctx) error {
	roles := c.Locals("roles").([]string)
	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to analyze tables",
		})
	}

	var payload dto.AnalyzeTablesRequest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid request payload",
		})
	}

	if err := h.validate.Struct(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	if err := h.service.AnalyzeTables(payload.TableIDs); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    nil,
		"message": "Tables analysis started successfully",
	})
}

func (h *TableHandler) CommitTable(c *fiber.Ctx) error {
	id := c.Params("id")
	roles := c.Locals("roles").([]string)
	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to commit table",
		})
	}

	if err := h.service.CommitTable(id); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    nil,
		"message": "Table committed successfully",
	})
}

func (h *TableHandler) CommitTables(c *fiber.Ctx) error {
	roles := c.Locals("roles").([]string)
	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to commit tables",
		})
	}

	var payload dto.CommitTablesRequest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid request payload",
		})
	}

	if err := h.validate.Struct(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	if err := h.service.CommitTables(payload.TableIDs); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    nil,
		"message": "Tables committed successfully",
	})
}

func (h *TableHandler) DownloadTable(c *fiber.Ctx) error {
	id := c.Params("id")
	roles := c.Locals("roles").([]string)
	orgID := c.Locals("organization_id").(*string)

	// Get multiple years from query parameter
	yearParams := c.Context().QueryArgs().PeekMulti("years")

	var years []int
	if len(yearParams) > 0 {
		// Parse each year parameter
		for _, yearParam := range yearParams {
			yearStr := string(yearParam)
			year := 0
			// Simple string to int conversion
			for _, ch := range yearStr {
				if ch >= '0' && ch <= '9' {
					year = year*10 + int(ch-'0')
				}
			}
			if year > 0 {
				years = append(years, year)
			}
		}
	}

	// If no years provided, default to last year
	if len(years) == 0 {
		years = []int{time.Now().Year() - 1}
	}

	// Get format from query parameter (default to xlsx)
	format := c.Query("format", "xlsx")

	data, err := h.service.DownloadTable(id, years, format, roles, orgID)
	if err != nil {
		status := 500
		if err == gorm.ErrRecordNotFound {
			status = 404
		}
		return c.Status(status).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	// Set Content-Type based on format
	contentType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	if format == "xls" {
		contentType = "application/vnd.ms-excel"
	}

	c.Set("Content-Disposition", "attachment; filename="+data.Name)
	c.Set("Content-Type", contentType)
	c.Status(200)
	return c.Send(data.File)
}

func (h *TableHandler) GenerateParentTable(c *fiber.Ctx) error {
	roles := c.Locals("roles").([]string)
	id := c.Params("id")
	var req dto.GenerateParentTableRequest

	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to generate parent table",
		})
	}

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid request body: " + err.Error(),
		})
	}

	// Validate request
	if err := h.validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"data":    nil,
			"message": "Validation error: " + err.Error(),
		})
	}

	// Generate parent table
	response, err := h.service.GenerateParentTable(id, &req)
	if err != nil {
		log.Printf("Error generating parent table: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"data":    nil,
			"message": "Failed to generate parent table: " + err.Error(),
		})
	}

	statusCode := fiber.StatusCreated
	if !response.IsNewTable {
		statusCode = fiber.StatusOK
	}

	return c.Status(statusCode).JSON(fiber.Map{
		"data":    response,
		"message": response.Message,
	})
}

func (h *TableHandler) UpdateTableIsIntegrated(c *fiber.Ctx) error {
	id := c.Params("id")
	roles := c.Locals("roles").([]string)

	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to update table integrate status",
		})
	}

	var payload dto.UpdateTableIsIntegratedRequest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid request payload",
		})
	}

	if err := h.validate.Struct(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	if err := h.service.UpdateTableIsIntegrated(id, payload.IsIntegrated); err != nil {
		status := 500
		if err == gorm.ErrRecordNotFound {
			status = 404
		}
		return c.Status(status).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    nil,
		"message": "Table is_integrated status updated successfully",
	})
}
