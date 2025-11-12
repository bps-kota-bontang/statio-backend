package dto

type UserResponse struct {
	ID             string   `json:"id"`
	Email          string   `json:"email"`
	OrganizationID *string  `json:"organization_id"`
	Roles          []string `json:"roles"`
}
