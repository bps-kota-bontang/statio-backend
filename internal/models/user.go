package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type User struct {
	ID                   string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Username             string         `gorm:"type:text;unique;not null"`
	Email                string         `gorm:"type:text;unique;not null"`
	Password             *string        `gorm:"type:text"`
	OrganizationID       *string        `gorm:"type:uuid;index"`
	Organization         *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	Roles                pq.StringArray `gorm:"type:text[]"`
	InviteToken          *string        `gorm:"type:text;uniqueIndex"`
	InviteTokenExpiresAt *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeletedAt            gorm.DeletedAt `gorm:"index"`
}
