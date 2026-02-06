package models

import (
	"time"

	"gorm.io/gorm"
)

type Dimension struct {
	ID        string           `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Code      *string          `gorm:"type:text;uniqueIndex"`
	Name      string           `gorm:"type:text;not null"`
	Notes     *string          `gorm:"type:text"`
	Aggregate bool             `gorm:"type:boolean;default:true"`
	Values    []DimensionValue `gorm:"foreignKey:DimensionID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
