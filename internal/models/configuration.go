package models

type Configuration struct {
	ID    string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name  string `gorm:"type:text;not null"`
	Key   string `gorm:"type:text;uniqueIndex"`
	Value string `gorm:"type:text;not null"`
}
