package handlers

import (
	"statio/internal/dto"
	"statio/internal/services"
	"statio/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ConfigurationHandler struct {
	service  *services.ConfigurationService
	validate *validator.Validate
}

func NewConfigurationHandler(service *services.ConfigurationService, validate *validator.Validate) *ConfigurationHandler {
	return &ConfigurationHandler{
		service:  service,
		validate: validate,
	}
}

func (h *ConfigurationHandler) GetConfiguration(c *fiber.Ctx) error {
	key := c.Params("key")
	roles := c.Locals("roles").([]string)

	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to get configuration",
		})
	}

	configuration, err := h.service.GetConfigurationByKey(key)
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
		"data":    configuration,
		"message": "Configuration fetched successfully",
	})
}

func (h *ConfigurationHandler) UpdateConfiguration(c *fiber.Ctx) error {
	key := c.Params("key")
	roles := c.Locals("roles").([]string)

	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to update configuration",
		})
	}

	var payload dto.UpdateConfigurationRequest
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

	if err := h.service.UpdateConfiguration(key, &payload); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    nil,
		"message": "Configuration updated successfully",
	})
}
