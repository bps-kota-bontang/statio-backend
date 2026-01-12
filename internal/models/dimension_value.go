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
	ParentID    *string              `gorm:"type:uuid;index"`
	Aggregate   *string              `gorm:"type:text;default:sum"` // e.g., "sum", "avg", etc.
	Dimension   *Dimension           `gorm:"foreignKey:DimensionID;constraint:OnDelete:CASCADE"`
	Parent      *DimensionValue      `gorm:"foreignKey:ParentID"`
	Children    []DimensionValue     `gorm:"foreignKey:ParentID"`
	FactValues  []FactDimensionValue `gorm:"foreignKey:DimensionValueID;constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
