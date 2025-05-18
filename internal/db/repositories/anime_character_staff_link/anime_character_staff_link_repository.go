package anime_character_staff_link

import (
	"github.com/weeb-vip/character-staff-sync/internal/db"
)

type AnimeCharacterStaffLinkRepositoryImpl interface {
}

type AnimeCharacterStaffLinkRepository struct {
	db *db.DB
}

func NewAnimeCharacterStaffLinkRepository(db *db.DB) AnimeCharacterStaffLinkRepositoryImpl {
	return &AnimeCharacterStaffLinkRepository{db: db}
}
