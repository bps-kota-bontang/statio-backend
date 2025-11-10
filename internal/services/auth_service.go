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
}

func NewAuthService(userService *UserService, jwtService *JWTService) *AuthService {
	return &AuthService{
		userService: userService,
		jwtService:  jwtService,
	}
}

func (s *AuthService) Login(payload *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userService.GetUserByEmail(payload.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// ✅ bandingkan hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// ✅ generate tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Roles)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// opsional: simpan refresh token ke DB
	// s.userService.SaveRefreshToken(user.ID, refreshToken)

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

	return s.jwtService.GenerateAccessToken(u.ID, u.Roles)
}
