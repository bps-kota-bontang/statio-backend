package dto

type LoginRequest struct {
	Identifier string `json:"identifier" validate:"required"` // can be email or username
	Password   string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LoginSSORequest struct {
	Code  string `json:"code" validate:"required"`
	State string `json:"state" validate:"required"`
	Type  string `json:"type" validate:"required"`
	Realm string `json:"realm" validate:"required"`
}

type LoginInviteTokenRequest struct {
	InviteToken string `json:"invite_token" validate:"required"`
}
