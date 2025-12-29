package repositories

import "statio/internal/models"

type UserRepository interface {
	FindByEmail(email string) (*models.User, error)
	FindByID(id string) (*models.User, error)
	FindByIDIncludePassword(id string) (*models.User, error)
	FindAll() ([]models.User, error)
	Count(search string, filters map[string][]string, total *int64) error
	FindPaginated(search string, limit, offset int, sortBy, sortOrder string, filters map[string][]string) ([]models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	Delete(id string) error
}
