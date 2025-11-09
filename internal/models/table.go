package models

import (
	"time"

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
	Dimensions     []TableDimension `gorm:"foreignKey:TableID;constraint:OnDelete:CASCADE"`
	Facts          []Fact           `gorm:"foreignKey:TableID;constraint:OnDelete:CASCADE"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}
