package anime_character_staff_link

import (
	"github.com/weeb-vip/character-staff-sync/internal/db"
	"gorm.io/gorm/clause"
)

type AnimeCharacterStaffLinkRepositoryImpl interface {
	Upsert(link *AnimeCharacterStaffLink) error
	Delete(link *AnimeCharacterStaffLink) error
}

type AnimeCharacterStaffLinkRepository struct {
	db *db.DB
}

func NewAnimeCharacterStaffLinkRepository(db *db.DB) AnimeCharacterStaffLinkRepositoryImpl {
	return &AnimeCharacterStaffLinkRepository{db: db}
}

func (r *AnimeCharacterStaffLinkRepository) Upsert(link *AnimeCharacterStaffLink) error {
	return r.db.DB.
		Clauses(clause.OnConflict{
			UpdateAll: true,
		}).
		Create(link).
		Error
}

func (r *AnimeCharacterStaffLinkRepository) Delete(link *AnimeCharacterStaffLink) error {
	return r.db.DB.Delete(link).Error
}
