package anime_character_staff_link

import (
	"github.com/weeb-vip/character-staff-sync/internal/db"
	"gorm.io/gorm/clause"
)

type AnimeCharacterStaffLinkRepository interface {
	Upsert(link *AnimeCharacterStaffLink) error
	Delete(link *AnimeCharacterStaffLink) error
}

type AnimeCharacterStaffLinkRepositoryImpl struct {
	db *db.DB
}

func NewAnimeCharacterStaffLinkRepository(db *db.DB) AnimeCharacterStaffLinkRepository {
	return &AnimeCharacterStaffLinkRepositoryImpl{db: db}
}

func (r *AnimeCharacterStaffLinkRepositoryImpl) Upsert(link *AnimeCharacterStaffLink) error {
	return r.db.DB.
		Clauses(clause.OnConflict{
			UpdateAll: true,
		}).
		Create(link).
		Error
}

func (r *AnimeCharacterStaffLinkRepositoryImpl) Delete(link *AnimeCharacterStaffLink) error {
	return r.db.DB.Delete(link).Error
}
