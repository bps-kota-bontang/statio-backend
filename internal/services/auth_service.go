package services

import (
	"errors"
	"fmt"
	"statio/internal/dto"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userService *UserService
	jwtService  *JWTService
	bpsService  *BPSService
}

func NewAuthService(userService *UserService, jwtService *JWTService, bpsService *BPSService) *AuthService {
	return &AuthService{
		userService: userService,
		jwtService:  jwtService,
		bpsService:  bpsService,
	}
}

func (s *AuthService) Login(payload *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userService.GetUserByEmailOrUsername(payload.Identifier)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// ✅ bandingkan hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// ✅ generate tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Roles, user.OrganizationID)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Refresh(refreshToken string) (string, error) {
	_, claims, err := s.jwtService.ParseToken(refreshToken)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}

	userID := claims["sub"].(string)
	u, err := s.userService.GetUserByID(userID)
	if err != nil {
		return "", errors.New("user not found")
	}

	return s.jwtService.GenerateAccessToken(u.ID, u.Roles, u.OrganizationID)
}

func (s *AuthService) LoginBPS(token string) (*dto.LoginResponse, error) {
	userInfo, err := s.bpsService.GetUserInfo(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	user, err := s.userService.GetUserByEmail(userInfo.Email)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// ✅ generate tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Roles, user.OrganizationID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) LoginInviteToken(inviteToken string) (*dto.LoginResponse, error) {
	user, err := s.userService.GetUserByInviteToken(inviteToken)
	if err != nil {
		return nil, fmt.Errorf("invalid invite token")
	}

	// ✅ generate tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Roles, user.OrganizationID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
