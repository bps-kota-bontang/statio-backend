package models

import (
	"time"

	"gorm.io/gorm"
)

type Project struct {
	ID        string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string         `gorm:"type:text;not null"`
	Tables    []ProjectTable `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ProjectTable struct {
	ProjectID   string   `gorm:"type:uuid;primaryKey"`
	TableID     string   `gorm:"type:uuid;primaryKey"`
	Page        string   `gorm:"type:text;not null"`
	TableNumber string   `gorm:"type:text;not null"`
	Project     *Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	Table       *Table   `gorm:"foreignKey:TableID;constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
