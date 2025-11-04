package models

import (
	"time"

	"gorm.io/gorm"
)

type FactDimensionValue struct {
	ID               string          `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	FactID           string          `gorm:"type:uuid;not null;index:idx_fact_dim"`
	DimensionValueID string          `gorm:"type:uuid;not null;index:idx_fact_dim"`
	Fact             *Fact           `gorm:"foreignKey:FactID;constraint:OnDelete:CASCADE"`
	DimensionValue   *DimensionValue `gorm:"foreignKey:DimensionValueID;constraint:OnDelete:CASCADE"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}
