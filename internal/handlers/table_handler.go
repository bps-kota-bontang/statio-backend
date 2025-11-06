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

	// Ambil filters per kolom, bisa multiple
	filters := map[string][]string{}
	keys := []string{
		"indicator_name",
		"indicator_measure",
		"indicator_unit",
		"dimensions",
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

	return c.JSON(fiber.Map{
		"data":    table,
		"message": "Table fetched successfully",
	})
}

func (h *TableHandler) UpdateFacts(c *fiber.Ctx) error {
	id := c.Params("id")

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

	if err := h.service.UpdateTableFacts(id, &payload); err != nil {
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
