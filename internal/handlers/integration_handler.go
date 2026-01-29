package handlers

import (
	"statio/internal/dto"
	"statio/internal/services"
	"statio/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type IntegrationHandler struct {
	service  *services.IntegrationService
	validate *validator.Validate
}

func NewIntegrationHandler(service *services.IntegrationService, validate *validator.Validate) *IntegrationHandler {
	return &IntegrationHandler{
		service:  service,
		validate: validate,
	}
}

func (h *IntegrationHandler) ExportDataIntegration(c *fiber.Ctx) error {
	roles := c.Locals("roles").([]string)
	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to export data integration",
		})
	}

	var payload dto.ExportDataIntegrationRequest
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

	data, err := h.service.ExportDataIntegration(payload.TableIDs, payload.Year)
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

	c.Set("Content-Disposition", "attachment; filename="+data.Name)
	c.Set("Content-Type", "application/zip")
	c.Status(200)
	return c.Send(data.File)
}
