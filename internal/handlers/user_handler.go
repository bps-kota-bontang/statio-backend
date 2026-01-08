package handlers

import (
	"statio/internal/dto"
	"statio/internal/services"
	"statio/utils"

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

func (h *UserHandler) UpdateMyEmail(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req dto.UpdateMyEmailRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid request payload",
		})
	}

	if err := h.validate.Struct(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	if err := h.service.UpdateMyEmail(userID, &req); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"data":    nil,
		"message": "Email updated successfully",
	})
}

func (h *UserHandler) UpdateMyPassword(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req dto.UpdateMyPasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid request payload",
		})
	}

	if err := h.validate.Struct(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	if err := h.service.UpdateMyPassword(userID, &req); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"data":    nil,
		"message": "Password updated successfully",
	})
}

func (h *UserHandler) GetAllUsers(c *fiber.Ctx) error {
	roles := c.Locals("roles").([]string)
	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to view users",
		})
	}

	sortBy := c.Query("sort_by", "no")
	sortOrder := c.Query("sort_order", "asc")
	search := c.Query("search")
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 10)

	// optional column filters
	filters := map[string][]string{}
	keys := []string{"roles", "organization_id"}
	for _, key := range keys {
		values := c.Context().QueryArgs().PeekMulti(key)
		if len(values) > 0 {
			strs := make([]string, len(values))
			for i, v := range values {
				strs[i] = string(v)
			}
			filters[key] = strs
		}
	}

	users, total, err := h.service.GetAllPaginated(search, page, perPage, sortBy, sortOrder, filters)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	meta := utils.NewPaginationMeta(total, page, perPage)

	return c.Status(200).JSON(fiber.Map{
		"data":    users,
		"message": "Users retrieved successfully",
		"meta":    meta,
	})
}

func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	id := c.Params("id")
	roles := c.Locals("roles").([]string)
	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to view users",
		})
	}

	user, err := h.service.GetUserByID(id)
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

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	roles := c.Locals("roles").([]string)
	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to create user",
		})
	}

	var req dto.CreateUserRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid request payload",
		})
	}

	if err := h.validate.Struct(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	if err := h.service.CreateUser(&req); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"data":    nil,
		"message": "User created successfully",
	})
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	roles := c.Locals("roles").([]string)

	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to  update user",
		})
	}

	id := c.Params("id")
	var req dto.UpdateUserRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid request payload",
		})
	}

	if err := h.validate.Struct(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	if err := h.service.UpdateUser(id, &req); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"data":    nil,
		"message": "User updated successfully",
	})
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	roles := c.Locals("roles").([]string)

	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to delete user",
		})
	}

	id := c.Params("id")

	if err := h.service.DeleteUser(id); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	c.Status(204)
	return nil
}

func (h *UserHandler) GetUserInviteLink(c *fiber.Ctx) error {
	roles := c.Locals("roles").([]string)
	if !utils.IsAdmin(roles) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"data":    nil,
			"message": "You are not authorized to view invite link",
		})
	}

	id := c.Params("id")

	inviteLinkResp, err := h.service.GetUserInviteLink(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"data":    inviteLinkResp,
		"message": "User invite link retrieved successfully",
	})
}
