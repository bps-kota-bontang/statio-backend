package handlers

import (
	"statio/internal/dto"
	"statio/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type OrganizationHandler struct {
	service  *services.OrganizationService
	validate *validator.Validate
}

func NewOrganizationHandler(service *services.OrganizationService, validate *validator.Validate) *OrganizationHandler {
	return &OrganizationHandler{service: service, validate: validate}
}

// Handler
func (h *OrganizationHandler) GetAllOrganizations(c *fiber.Ctx) error {
	organizations, err := h.service.GetAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    organizations,
		"message": "Organizations fetched successfully",
	})
}

// CreateOrganizationTable handles the creation of a new table for a specific organization
func (h *OrganizationHandler) AssignTables(c *fiber.Ctx) error {
	id := c.Params("id")

	var payload dto.AssignTablesRequest
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

	if err := h.service.AssignTablesToOrganization(id, &payload); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    nil,
		"message": "Table created successfully for organization",
	})
}
