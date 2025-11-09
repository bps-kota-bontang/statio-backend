package handlers

import (
	"statio/internal/dto"
	"statio/internal/services"
	"statio/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type DimensionHandler struct {
	service  *services.DimensionService
	validate *validator.Validate
}

func NewDimensionHandler(service *services.DimensionService, validate *validator.Validate) *DimensionHandler {
	return &DimensionHandler{service: service, validate: validate}
}

// Handler
func (h *DimensionHandler) GetAllDimensions(c *fiber.Ctx) error {
	sortBy := c.Query("sort_by", "no")
	sortOrder := c.Query("sort_order", "asc")
	search := c.Query("search")
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 10)

	// Ambil filters per kolom, bisa multiple
	filters := map[string][]string{}
	keys := []string{"name"}
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

func (h *DimensionHandler) GetDimension(c *fiber.Ctx) error {
	id := c.Params("id")

	Dimension, err := h.service.GetByID(id)
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
		"data":    Dimension,
		"message": "Dimension fetched successfully",
	})
}

func (h *DimensionHandler) CreateDimension(c *fiber.Ctx) error {
	var payload dto.CreateDimensionRequest
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

	Dimension, err := h.service.Create(&payload)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"data":    Dimension,
		"message": "Dimension created successfully",
	})
}
func (h *DimensionHandler) UpdateDimension(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload dto.UpdateDimensionRequest
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

	Dimension, err := h.service.Update(id, &payload)
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
		"data":    Dimension,
		"message": "Dimension updated successfully",
	})
}

func (h *DimensionHandler) GetAllDimensionNames(c *fiber.Ctx) error {
	Dimensions, err := h.service.GetAllNames()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    Dimensions,
		"message": "Dimension names fetched successfully",
	})
}
