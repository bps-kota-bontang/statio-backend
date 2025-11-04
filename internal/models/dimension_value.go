package models

import (
	"time"

	"gorm.io/gorm"
)

type DimensionValue struct {
	ID          string               `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	DimensionID string               `gorm:"type:uuid;not null;index"`
	Name        string               `gorm:"type:text;not null"`
	Order       int                  `gorm:"type:int;default:0"`
	Dimension   *Dimension           `gorm:"foreignKey:DimensionID;constraint:OnDelete:CASCADE"`
	FactValues  []FactDimensionValue `gorm:"foreignKey:DimensionValueID;constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
