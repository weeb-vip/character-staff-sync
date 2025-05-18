package anime_character_staff_link

import (
	"time"
)

type AnimeCharacterStaffLink struct {
	ID              string    `gorm:"type:char(36);primaryKey"`
	CharacterID     string    `gorm:"type:varchar(36);not null"`
	StaffID         string    `gorm:"type:varchar(36);not null"`
	CharacterName   string    `gorm:"type:varchar(255);not null"`
	StaffGivenName  string    `gorm:"type:varchar(255);not null"`
	StaffFamilyName string    `gorm:"type:varchar(255);not null"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}

func (AnimeCharacterStaffLink) TableName() string {
	return "anime_character_staff_link"
}
