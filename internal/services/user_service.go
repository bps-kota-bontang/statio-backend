package services

import (
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

func (s *UserService) GetUserByID(id string) (*models.User, error) {
	return s.userRepo.FindByID(id)
}
