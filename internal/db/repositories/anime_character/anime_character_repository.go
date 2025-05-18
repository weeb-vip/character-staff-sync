package anime_character

import (
	"github.com/weeb-vip/character-staff-sync/internal/db"
	"gorm.io/gorm/clause"
)

type AnimeCharacterRepositoryImpl interface {
	Upsert(character *AnimeCharacter) error
	Delete(character *AnimeCharacter) error
	FindByID(id string) (*AnimeCharacter, error)
	FindByName(name string) (string, error)
}

type AnimeCharacterRepository struct {
	db *db.DB
}

func NewAnimeCharacterRepository(db *db.DB) AnimeCharacterRepositoryImpl {
	return &AnimeCharacterRepository{db: db}
}

func (r *AnimeCharacterRepository) Upsert(character *AnimeCharacter) error {
	return r.db.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(character).Error
}

func (r *AnimeCharacterRepository) Delete(character *AnimeCharacter) error {
	return r.db.DB.Delete(character).Error
}

func (r *AnimeCharacterRepository) FindByID(id string) (*AnimeCharacter, error) {
	var result AnimeCharacter
	err := r.db.DB.First(&result, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *AnimeCharacterRepository) FindByName(name string) (string, error) {
	var character AnimeCharacter
	err := r.db.DB.Select("id").Where("name = ?", name).First(&character).Error
	if err != nil {
		return "", err
	}
	return character.ID, nil
}
