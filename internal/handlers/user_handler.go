package handlers

import (
	"statio/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	service  *services.UserService
	validate *validator.Validate
}

func NewUserHandler(service *services.UserService, validate *validator.Validate) *UserHandler {
	return &UserHandler{
		service:  service,
		validate: validate,
	}
}

func (h *UserHandler) Me(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	user, err := h.service.GetUserByID(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"data":    user,
		"message": "User retrieved successfully",
	})
}
