package mappers

import (
	"statio/internal/dto"
	"statio/internal/models"
)

func ToUserResponse(user *models.User) *dto.UserResponse {
	resp := dto.UserResponse{
		ID:             user.ID,
		Email:          user.Email,
		OrganizationID: user.OrganizationID,
		Roles:          user.Roles,
	}

	if user.Organization != nil {
		resp.Organization = ToOrganizationResponse(user.Organization)
	}

	return &resp
}

func ToUserModel(input *dto.CreateUserRequest) *models.User {
	return &models.User{
		Email:          input.Email,
		OrganizationID: &input.OrganizationID,
		Roles:          input.Roles,
	}
}

func ApplyUserUpdateFromRequest(user *models.User, req *dto.UpdateUserRequest) {
	if req.Email != nil {
		user.Email = *req.Email
	}

	if req.OrganizationID != nil {
		user.OrganizationID = req.OrganizationID
	}

	if req.Roles != nil {
		user.Roles = *req.Roles
	}
}
