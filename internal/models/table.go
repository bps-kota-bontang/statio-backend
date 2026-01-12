package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Table struct {
	ID             string           `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name           string           `gorm:"type:text;not null"`
	Direction      int              `gorm:"type:smallint;not null;default:1"`
	Description    *string          `gorm:"type:text"`
	IndicatorID    string           `gorm:"type:uuid;index"`
	Indicator      *Indicator       `gorm:"foreignKey:IndicatorID;constraint:OnDelete:CASCADE"`
	OrganizationID *string          `gorm:"type:uuid;index"`
	Organization   *Organization    `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	Labels         pq.StringArray   `gorm:"type:text[]"`
	Notes          *string          `gorm:"type:text"`
	Aggregate      *string          `gorm:"type:text;default:sum"` // Rumus agregasi dalam format teks
	IsLocked       bool             `gorm:"type:boolean;not null;default:false"`
	Status         string           `gorm:"type:text;not null;default:'draft'"`
	IsAggregated   bool             `gorm:"type:boolean;not null;default:false;index"` // Menandai table hasil agregasi
	SourceTableID  *string          `gorm:"type:uuid;index"`                           // ID table sumber untuk table agregasi
	Dimensions     []TableDimension `gorm:"foreignKey:TableID;constraint:OnDelete:CASCADE"`
	Facts          []Fact           `gorm:"foreignKey:TableID;constraint:OnDelete:CASCADE"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}
