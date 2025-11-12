package mappers

import (
	"statio/internal/dto"
	"statio/internal/models"
)

func ToUserResponse(user *models.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:             user.ID,
		Email:          user.Email,
		OrganizationID: user.OrganizationID,
		Roles:          user.Roles,
	}
}
