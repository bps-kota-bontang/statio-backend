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
	Token string `json:"token" validate:"required"`
	State string `json:"state" validate:"required"`
}
