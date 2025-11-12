package handlers

import (
	"statio/internal/dto"
	"statio/internal/services"
	"statio/utils"

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

func (h *OrganizationHandler) AssignTables(c *fiber.Ctx) error {
	id := c.Params("id")
	roles := c.Locals("roles").([]string)

	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to create organizations",
		})
	}

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

func (h *OrganizationHandler) CreateOrganization(c *fiber.Ctx) error {
	roles := c.Locals("roles").([]string)

	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to create organizations",
		})
	}

	var payload dto.CreateOrganizationRequest
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
	resp, err := h.service.Create(&payload)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    resp,
		"message": "Organization created successfully",
	})
}

func (h *OrganizationHandler) UpdateOrganization(c *fiber.Ctx) error {
	id := c.Params("id")
	roles := c.Locals("roles").([]string)

	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to create organizations",
		})
	}

	var payload dto.UpdateOrganizationRequest
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

	org, err := h.service.Update(id, &payload)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    org,
		"message": "Organization updated successfully",
	})
}
