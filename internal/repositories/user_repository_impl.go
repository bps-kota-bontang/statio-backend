package repositories

import (
	"statio/internal/models"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type UserRepositoryImpl struct {
	db *gorm.DB
}

// FindByIDIncludePassword implements UserRepository.
func (u *UserRepositoryImpl) FindByIDIncludePassword(id string) (*models.User, error) {
	var user *models.User
	if err := u.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// Create implements UserRepository.
func (u *UserRepositoryImpl) Create(user *models.User) error {
	return u.db.Create(user).Error
}

// Delete implements UserRepository.
func (u *UserRepositoryImpl) Delete(id string) error {
	return u.db.Delete(&models.User{}, "id = ?", id).Error
}

// Update implements UserRepository.
func (u *UserRepositoryImpl) Update(user *models.User) error {
	return u.db.Save(user).Error
}

// FindAll implements UserRepository.
func (u *UserRepositoryImpl) FindAll() ([]models.User, error) {
	var users []models.User
	if err := u.db.Preload("Organization").Omit("password").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
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
	if err := u.db.Omit("password").Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{db: db}
}

// Count implements UserRepository.Count
func (u *UserRepositoryImpl) Count(search string, filters map[string][]string, total *int64) error {
	query := u.db.Model(&models.User{})

	if search != "" {
		like := "%" + search + "%"
		query = query.Where("email ILIKE ?", like)
	}

	// apply simple filters (supports __NULL__ marker and IN semantics)
	for col, values := range filters {
		if len(values) == 0 {
			continue
		}

		hasNull := false
		realValues := make([]string, 0, len(values))
		for _, v := range values {
			if v == "__NULL__" {
				hasNull = true
			} else {
				realValues = append(realValues, v)
			}
		}

		if col == "roles" {
			// roles is a Postgres text[] column — use overlap (&&) operator
			if hasNull && len(realValues) > 0 {
				query = query.Where("("+col+" && ? OR "+col+" IS NULL OR "+col+" = '{}')", pq.Array(realValues))
			} else if hasNull {
				query = query.Where(col + " IS NULL OR " + col + " = '{}'")
			} else {
				query = query.Where(col+" && ?", pq.Array(realValues))
			}
		} else {
			if hasNull && len(realValues) > 0 {
				query = query.Where("("+col+" IN ? OR "+col+" IS NULL OR "+col+" = '')", realValues)
			} else if hasNull {
				query = query.Where(col + " IS NULL OR " + col + " = ''")
			} else {
				query = query.Where(col+" IN ?", realValues)
			}
		}
	}

	return query.Distinct("users.id").Count(total).Error
}

// FindPaginated implements UserRepository.FindPaginated
func (u *UserRepositoryImpl) FindPaginated(search string, limit, offset int, sortBy, sortOrder string, filters map[string][]string) ([]models.User, error) {
	query := u.db.Model(&models.User{}).Preload("Organization").Omit("password")

	if search != "" {
		like := "%" + search + "%"
		query = query.Where("email ILIKE ?", like)
	}

	for col, values := range filters {
		if len(values) == 0 {
			continue
		}

		hasNull := false
		realValues := make([]string, 0, len(values))
		for _, v := range values {
			if v == "__NULL__" {
				hasNull = true
			} else {
				realValues = append(realValues, v)
			}
		}

		if col == "roles" {
			if hasNull && len(realValues) > 0 {
				query = query.Where("("+col+" && ? OR "+col+" IS NULL OR "+col+" = '{}')", pq.Array(realValues))
			} else if hasNull {
				query = query.Where(col + " IS NULL OR " + col + " = '{}'")
			} else {
				query = query.Where(col+" && ?", pq.Array(realValues))
			}
		} else {
			if hasNull && len(realValues) > 0 {
				query = query.Where("("+col+" IN ? OR "+col+" IS NULL OR "+col+" = '')", realValues)
			} else if hasNull {
				query = query.Where(col + " IS NULL OR " + col + " = ''")
			} else {
				query = query.Where(col+" IN ?", realValues)
			}
		}
	}

	validSortFields := map[string]string{
		"no":    "created_at",
		"email": "email",
	}
	field, ok := validSortFields[sortBy]
	if !ok {
		field = "created_at"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}

	var users []models.User
	if err := query.Order(field + " " + sortOrder).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
