package services

import (
	"statio/internal/dto"
	"statio/internal/mappers"
	"statio/internal/models"
	"statio/internal/repositories"
)

type UserService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	return s.userRepo.FindByEmail(email)
}

func (s *UserService) GetUserByID(id string) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	resp := mappers.ToUserResponse(user)
	return resp, nil
}
