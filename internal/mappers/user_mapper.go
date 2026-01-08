package mappers

import (
	"statio/internal/dto"
	"statio/internal/models"
)

func ToUserResponse(user *models.User) *dto.UserResponse {
	resp := dto.UserResponse{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		OrganizationID: user.OrganizationID,
		Roles:          user.Roles,
		HasInviteLink:  user.InviteToken != nil,
		HasPassword:    user.Password != nil,
	}

	if user.Organization != nil {
		resp.Organization = ToOrganizationResponse(user.Organization)
	}

	return &resp
}

func ToUserModel(input *dto.CreateUserRequest) *models.User {
	return &models.User{
		Username:       input.Username,
		Email:          input.Email,
		OrganizationID: input.OrganizationID,
		Roles:          input.Roles,
		InviteToken:    input.InviteToken,
	}
}

func ApplyUserUpdateFromRequest(user *models.User, req *dto.UpdateUserRequest) {
	user.Username = req.Username
	user.Email = req.Email
	user.OrganizationID = req.OrganizationID
	user.Roles = req.Roles

	if req.InviteToken != nil {
		user.InviteToken = req.InviteToken
	}
}
