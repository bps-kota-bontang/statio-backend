package models

import (
	"time"

	"gorm.io/gorm"
)

type Fact struct {
	ID                  string               `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TableID             string               `gorm:"type:uuid;not null;index:idx_table_year"` // tambahkan nama index
	Value               *float64             `gorm:"type:double precision"`
	Year                int                  `gorm:"not null;index:idx_table_year"` // ikut index yang sama
	Table               *Table               `gorm:"foreignKey:TableID;constraint:OnDelete:CASCADE"`
	FactDimensionValues []FactDimensionValue `gorm:"foreignKey:FactID;constraint:OnDelete:CASCADE"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           gorm.DeletedAt `gorm:"index"`
}
