package providers

import (
	"errors"
	"fmt"
	"log"
	"statio/config"
	"statio/internal/models"
	"strings"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewDBConnection establishes a connection to the database based on provided config and returns the DB instance
func NewDBConnection(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	// Initialize the DSN (Data Source Name) and Dialector based on the DB driver
	var dsn string
	var dialector gorm.Dialector

	switch cfg.DBDriver {
	case "postgres":
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)
		dialector = postgres.Open(dsn)

	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
		dialector = mysql.Open(dsn)

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.DBDriver)
	}

	// Open the DB connection
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
		return nil, err
	}

	// Run database migrations
	if err := db.AutoMigrate(
		&models.Table{},
		&models.Indicator{},
		&models.Dimension{},
		&models.DimensionValue{},
		//&models.TableIndicator{},
		&models.TableDimension{},
		&models.Fact{},
		&models.FactDimensionValue{},
		&models.Organization{},
		&models.User{},
		&models.Configuration{},
	); err != nil {
		log.Fatalf("Failed to apply database migrations: %v", err)
		return nil, err
	}

	log.Println("Database connection established successfully")
	return db, nil
}

func buildSeedAdminUser(username, email, password string) *models.User {
	normalizedUsername := strings.TrimSpace(username)
	if normalizedUsername == "" {
		normalizedUsername = "dataadmin"
	}

	normalizedEmail := strings.TrimSpace(email)
	if normalizedEmail == "" {
		normalizedEmail = "dataadmin@statio.local"
	}

	normalizedPassword := strings.TrimSpace(password)
	if normalizedPassword == "" {
		normalizedPassword = "dataadmin123"
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(normalizedPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil
	}

	passwordStr := string(hashedPassword)
	return &models.User{
		Username: normalizedUsername,
		Email:    &normalizedEmail,
		Password: &passwordStr,
		Roles:    pq.StringArray{"admin"},
	}
}

func SeedDataAdmin(db *gorm.DB, username, email, password string) (*models.User, error) {
	seedUser := buildSeedAdminUser(username, email, password)
	if seedUser == nil {
		return nil, fmt.Errorf("failed to build seed admin user")
	}

	var existing models.User
	err := db.Where("username ILIKE ? OR email ILIKE ?", seedUser.Username, *seedUser.Email).First(&existing).Error
	if err == nil {
		existing.Username = seedUser.Username
		existing.Email = seedUser.Email
		existing.Password = seedUser.Password
		existing.Roles = seedUser.Roles
		if saveErr := db.Save(&existing).Error; saveErr != nil {
			return nil, saveErr
		}
		return &existing, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if createErr := db.Create(seedUser).Error; createErr != nil {
		return nil, createErr
	}

	return seedUser, nil
}
