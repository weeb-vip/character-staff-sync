package anime_staff

import (
	"github.com/weeb-vip/character-staff-sync/internal/db"
	"gorm.io/gorm/clause"
)

type AnimeStaffRepositoryImpl interface {
	Upsert(staff *AnimeStaff) error
	Delete(staff *AnimeStaff) error
	FindByID(id string) (*AnimeStaff, error) // Optional but helpful
	FindByFullName(givenName string, familyName string) (string, error)
}

type AnimeStaffRepository struct {
	db *db.DB
}

func NewAnimeStaffRepository(db *db.DB) AnimeStaffRepositoryImpl {
	return &AnimeStaffRepository{db: db}
}

func (r *AnimeStaffRepository) Upsert(staff *AnimeStaff) error {
	return r.db.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(staff).Error
}

func (r *AnimeStaffRepository) Delete(staff *AnimeStaff) error {
	return r.db.DB.Delete(staff).Error
}

func (r *AnimeStaffRepository) FindByID(id string) (*AnimeStaff, error) {
	var result AnimeStaff
	err := r.db.DB.First(&result, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *AnimeStaffRepository) FindByFullName(givenName string, familyName string) (string, error) {
	var staff AnimeStaff
	err := r.db.DB.Select("id").
		Where("given_name = ? AND family_name = ?", givenName, familyName).
		First(&staff).Error
	if err != nil {
		return "", err
	}
	return staff.ID, nil
}
