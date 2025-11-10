package repositories

import "statio/internal/models"

type UserRepository interface {
	FindByEmail(email string) (*models.User, error)
	FindByID(id string) (*models.User, error)
}
