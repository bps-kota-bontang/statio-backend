package handlers

import (
	"statio/internal/dto"
	"statio/internal/services"
	"statio/utils"

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

	if !utils.IsAdmin(roles) {
		if organizationID != nil {
			filters["organization_id"] = []string{*organizationID}
		} else {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"data":    nil,
				"message": "Missing organization context for operator access",
			})
		}
	}

	Dimensions, total, err := h.service.GetAllPaginated(search, page, perPage, sortBy, sortOrder, filters)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	meta := utils.NewPaginationMeta(total, page, perPage)

	return c.JSON(fiber.Map{
		"data":    Dimensions,
		"message": "Dimensions fetched successfully",
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

	if !utils.IsAdmin(roles) {
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

func (h *TableHandler) CreateTable(c *fiber.Ctx) error {
	var payload dto.CreateTableRequest
	roles := c.Locals("roles").([]string)

	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to create tables",
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
			"message": "You are not authorized to create tables",
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
