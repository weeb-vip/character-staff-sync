package anime_staff

import (
	"time"
)

type AnimeStaff struct {
	ID         string    `gorm:"type:char(36);primaryKey"`
	GivenName  string    `gorm:"type:varchar(255);not null"`
	FamilyName string    `gorm:"type:varchar(255);not null"`
	Image      string    `gorm:"type:text"`
	Birthday   string    `gorm:"type:varchar(255)"`
	BirthPlace string    `gorm:"type:varchar(255)"`
	BloodType  string    `gorm:"type:varchar(255)"`
	Hobbies    string    `gorm:"type:varchar(255)"`
	Summary    string    `gorm:"type:text"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}

func (AnimeStaff) TableName() string {
	return "anime_staff"
}
