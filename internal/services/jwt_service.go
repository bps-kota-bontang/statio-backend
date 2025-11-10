package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type AccessClaims struct {
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

func NewJWTService(secret string, accessTTL, refreshTTL time.Duration) *JWTService {
	return &JWTService{
		Secret:          secret,
		AccessTokenTTL:  accessTTL,
		RefreshTokenTTL: refreshTTL,
	}
}

func (m *JWTService) GenerateAccessToken(userID string, roles []string) (string, error) {
	claims := AccessClaims{
		UserID: userID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.Secret))
}

func (m *JWTService) GenerateRefreshToken(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.RefreshTokenTTL)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.Secret))
}

func (m *JWTService) ParseToken(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte(m.Secret), nil
	})
	if err != nil || !token.Valid {
		return nil, nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, jwt.ErrTokenInvalidClaims
	}

	return token, claims, nil
}

func (j *JWTService) ValidateAccessToken(tokenString string) (string, []string, error) {
	_, claims, err := j.ParseToken(tokenString)
	if err != nil {
		return "", nil, errors.New("invalid or expired token")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", nil, errors.New("invalid token payload")
	}

	roles := make([]string, 0)
	if r, ok := claims["roles"].([]interface{}); ok {
		for _, v := range r {
			if str, ok := v.(string); ok {
				roles = append(roles, str)
			}
		}
	}

	return userID, roles, nil
}
