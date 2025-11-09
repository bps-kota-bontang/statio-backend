package models

import (
	"time"

	"gorm.io/gorm"
)

type Organization struct {
	ID        string  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string  `gorm:"type:text;not null;uniqueIndex"`
	Tables    []Table `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
