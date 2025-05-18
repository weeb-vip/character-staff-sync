package anime_character

import (
	"github.com/weeb-vip/character-staff-sync/internal/db"
)

type AnimeCharacterRepositoryImpl interface {
}

type AnimeCharacterRepository struct {
	db *db.DB
}

func NewAnimeCharacterRepository(db *db.DB) AnimeCharacterRepositoryImpl {
	return &AnimeCharacterRepository{db: db}
}
