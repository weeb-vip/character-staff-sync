package pulsar_anime_character_staff_link_postgres_processor

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/weeb-vip/character-staff-sync/internal/db"
	"github.com/weeb-vip/character-staff-sync/internal/db/repositories/anime_character"
	"github.com/weeb-vip/character-staff-sync/internal/db/repositories/anime_character_staff_link"
	"github.com/weeb-vip/character-staff-sync/internal/db/repositories/anime_staff"
	"github.com/weeb-vip/character-staff-sync/internal/producer"
)

type Options struct {
	NoErrorOnDelete bool
}

type PulsarAnimeCharacterStaffLinkPostgresProcessorImpl interface {
	Process(ctx context.Context, data Payload) error
}

type PulsarAnimeCharacterStaffLinkPostgresProcessor struct {
	LinkRepo  anime_character_staff_link.AnimeCharacterStaffLinkRepositoryImpl
	CharRepo  anime_character.AnimeCharacterRepositoryImpl
	StaffRepo anime_staff.AnimeStaffRepositoryImpl
	Options   Options
	Producer  producer.Producer[Schema]
}

func NewPulsarAnimeCharacterStaffLinkPostgresProcessor(opt Options, db *db.DB, prod producer.Producer[Schema]) PulsarAnimeCharacterStaffLinkPostgresProcessorImpl {
	return &PulsarAnimeCharacterStaffLinkPostgresProcessor{
		LinkRepo:  anime_character_staff_link.NewAnimeCharacterStaffLinkRepository(db),
		CharRepo:  anime_character.NewAnimeCharacterRepository(db),
		StaffRepo: anime_staff.NewAnimeStaffRepository(db),
		Options:   opt,
		Producer:  prod,
	}
}

func (p *PulsarAnimeCharacterStaffLinkPostgresProcessor) Process(ctx context.Context, data Payload) error {
	if data.After == nil {
		log.Println("WARN: data.After is nil, skipping link processing")
		return nil
	}

	charId, err := p.CharRepo.FindByName(*data.After.CharacterName)
	if err != nil {
		return err
	}

	staffId, err := p.StaffRepo.FindByFullName(*data.After.StaffGivenName, *data.After.StaffFamilyName)
	if err != nil {
		return err
	}

	link := &anime_character_staff_link.AnimeCharacterStaffLink{
		CharacterID:     charId,
		StaffID:         staffId,
		CharacterName:   ptrToString(data.After.CharacterName),
		StaffGivenName:  ptrToString(data.After.StaffGivenName),
		StaffFamilyName: ptrToString(data.After.StaffFamilyName),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err = p.LinkRepo.Upsert(link)
	if err != nil {
		return err
	}

	jsonLink, err := json.Marshal(ProducerPayload{
		Action: CreateAction,
		Data:   data.After,
	})
	if err != nil {
		return err
	}

	return p.Producer.Send(ctx, jsonLink)
}

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
