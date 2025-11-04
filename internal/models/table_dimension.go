package models

import (
	"time"

	"gorm.io/gorm"
)

type TableDimension struct {
	ID          string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TableID     string     `gorm:"type:uuid;not null;index:idx_table_dim"`
	DimensionID string     `gorm:"type:uuid;not null;index:idx_table_dim"`
	Table       *Table     `gorm:"foreignKey:TableID;constraint:OnDelete:CASCADE"`
	Dimension   *Dimension `gorm:"foreignKey:DimensionID;constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
