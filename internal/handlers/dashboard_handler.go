package handlers

import (
	"statio/internal/services"
	"statio/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type DashboardHandler struct {
	service  *services.DashboardService
	validate *validator.Validate
}

func NewDashboardHandler(service *services.DashboardService, validate *validator.Validate) *DashboardHandler {
	return &DashboardHandler{
		service:  service,
		validate: validate,
	}
}

func (h *DashboardHandler) GetDashboardStatistics(c *fiber.Ctx) error {
	roles := c.Locals("roles").([]string)
	orgID := c.Locals("organization_id").(*string)

	if utils.IsAdmin(roles) {
		orgID = nil
	}

	stats, err := h.service.GetDashboardStatistics(orgID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    stats,
		"message": "Dashboard statistics fetched successfully",
	})
}
