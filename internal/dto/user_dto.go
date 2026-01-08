package dto

type UserResponse struct {
	ID             string                `json:"id"`
	Username       string                `json:"username"`
	Email          *string               `json:"email"`
	OrganizationID *string               `json:"organization_id"`
	Roles          []string              `json:"roles"`
	Organization   *OrganizationResponse `json:"organization,omitempty"`
	HasInviteLink  bool                  `json:"has_invite_link"`
	HasPassword    bool                  `json:"has_password"`
}

type UserInviteLinkResponse struct {
	InviteLink string `json:"invite_link"`
}

type CreateUserRequest struct {
	Username       string   `json:"username" validate:"required"`
	Email          *string  `json:"email,omitempty" validate:"omitempty,email"`
	Password       *string  `json:"password,omitempty"`
	Roles          []string `json:"roles" validate:"required,min=1,dive,required"`
	OrganizationID *string  `json:"organization_id,omitempty"`
	InviteToken    *string  `json:"invite_token,omitempty"`
}

type UpdateUserRequest struct {
	Username       string   `json:"username" validate:"required"`
	Email          *string  `json:"email,omitempty" validate:"omitempty,email"`
	Password       *string  `json:"password,omitempty"`
	Roles          []string `json:"roles" validate:"required,min=1,dive,required"`
	OrganizationID *string  `json:"organization_id,omitempty"`
	InviteToken    *string  `json:"invite_token,omitempty"`
}

type UpdateMyEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type UpdateMyPasswordRequest struct {
	OldPassword     *string `json:"old_password,omitempty"`
	NewPassword     string  `json:"new_password" validate:"required"`
	ConfirmPassword string  `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}
