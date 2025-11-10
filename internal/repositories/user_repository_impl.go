package repositories

import (
	"statio/internal/models"

	"gorm.io/gorm"
)

type UserRepositoryImpl struct {
	db *gorm.DB
}

// FindByEmail implements UserRepository.
func (u *UserRepositoryImpl) FindByEmail(email string) (*models.User, error) {
	var user *models.User
	if err := u.db.Where("email ILIKE ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// FindByID implements UserRepository.
func (u *UserRepositoryImpl) FindByID(id string) (*models.User, error) {
	var user *models.User
	if err := u.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{db: db}
}
