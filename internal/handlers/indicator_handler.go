package handlers

import (
	"statio/internal/dto"
	"statio/internal/services"
	"statio/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type IndicatorHandler struct {
	service  *services.IndicatorService
	validate *validator.Validate
}

func NewIndicatorHandler(service *services.IndicatorService, validate *validator.Validate) *IndicatorHandler {
	return &IndicatorHandler{service: service, validate: validate}
}

func (h *IndicatorHandler) GetAllIndicators(c *fiber.Ctx) error {
	sortBy := c.Query("sort_by", "no")
	sortOrder := c.Query("sort_order", "asc")
	search := c.Query("search")
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 10)

	// Ambil filters per kolom, bisa multiple
	filters := map[string][]string{}
	keys := []string{"measure", "unit"}
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

	indicators, total, err := h.service.GetAllPaginated(search, page, perPage, sortBy, sortOrder, filters)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	meta := utils.NewPaginationMeta(total, page, perPage)

	return c.JSON(fiber.Map{
		"data":    indicators,
		"message": "Indicators fetched successfully",
		"meta":    meta,
	})
}

func (h *IndicatorHandler) GetIndicator(c *fiber.Ctx) error {
	id := c.Params("id")

	indicator, err := h.service.GetByID(id)
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
		"data":    indicator,
		"message": "Indicator fetched successfully",
	})
}

func (h *IndicatorHandler) CreateIndicator(c *fiber.Ctx) error {
	roles := c.Locals("roles").([]string)

	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to create indicator",
		})
	}

	var payload dto.CreateIndicatorRequest
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

	indicator, err := h.service.Create(&payload)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"data":    indicator,
		"message": "Indicator created successfully",
	})
}

func (h *IndicatorHandler) UpdateIndicator(c *fiber.Ctx) error {
	id := c.Params("id")
	roles := c.Locals("roles").([]string)

	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to update indicator",
		})
	}

	var payload dto.UpdateIndicatorRequest
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

	indicator, err := h.service.Update(id, &payload)
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
		"data":    indicator,
		"message": "Indicator updated successfully",
	})
}

func (h *IndicatorHandler) GetAllIndicatorNames(c *fiber.Ctx) error {
	names, err := h.service.GetAllNames()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    names,
		"message": "Indicator names fetched successfully",
	})
}

func (h *IndicatorHandler) GetAllIndicatorMeasures(c *fiber.Ctx) error {
	measures, err := h.service.GetAllMeasures()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    measures,
		"message": "Indicator measures fetched successfully",
	})
}

func (h *IndicatorHandler) GetAllIndicatorUnits(c *fiber.Ctx) error {
	units, err := h.service.GetAllUnits()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    units,
		"message": "Indicator units fetched successfully",
	})
}
