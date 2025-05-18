package anime_character

import (
	"time"
)

type AnimeCharacter struct {
	ID            string    `gorm:"type:char(36);primaryKey"`
	AnimeID       string    `gorm:"type:varchar(36);not null"`
	Name          string    `gorm:"type:varchar(255);not null"`
	Role          string    `gorm:"type:varchar(255);not null"`
	Birthday      string    `gorm:"type:varchar(255)"`
	Zodiac        string    `gorm:"type:varchar(255)"`
	Gender        string    `gorm:"type:varchar(255)"`
	Race          string    `gorm:"type:varchar(255)"`
	Height        string    `gorm:"type:varchar(255)"`
	Weight        string    `gorm:"type:varchar(255)"`
	Title         string    `gorm:"type:varchar(255)"`
	MartialStatus string    `gorm:"type:varchar(255)"`
	Summary       string    `gorm:"type:text"`
	Image         string    `gorm:"type:text"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

func (AnimeCharacter) TableName() string {
	return "anime_character"
}
