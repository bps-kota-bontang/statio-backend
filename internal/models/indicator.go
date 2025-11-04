package models

import (
	"time"

	"gorm.io/gorm"
)

type Indicator struct {
	ID          string  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Code        *string `gorm:"type:text;uniqueIndex"`
	Name        string  `gorm:"type:text;not null"`
	Description *string `gorm:"type:text"`
	Measure     string  `gorm:"type:text;not null"` // Total, Rata-rata, Persentase, Indeks, Nilai, dll
	Unit        *string `gorm:"type:text"`          // Orang, Rupiah, Kilometer, dll
	Tables      []Table `gorm:"foreignKey:IndicatorID;constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
