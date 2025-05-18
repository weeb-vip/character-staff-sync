package pulsar_anime_character_postgres_processor

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/weeb-vip/character-staff-sync/internal/db"
	"github.com/weeb-vip/character-staff-sync/internal/db/repositories/anime_character"
	"github.com/weeb-vip/character-staff-sync/internal/producer"
)

type Options struct {
	NoErrorOnDelete bool
}

type PulsarAnimeCharacterPostgresProcessorImpl interface {
	Process(ctx context.Context, data Payload) error
	parseToEntity(ctx context.Context, data Schema) (*anime_character.AnimeCharacter, error)
}

type PulsarAnimeCharacterPostgresProcessor struct {
	Repository anime_character.AnimeCharacterRepositoryImpl
	Options    Options
	Producer   producer.Producer[Schema]
}

func NewPulsarAnimeCharacterPostgresProcessor(opt Options, db *db.DB, prod producer.Producer[Schema]) PulsarAnimeCharacterPostgresProcessorImpl {
	return &PulsarAnimeCharacterPostgresProcessor{
		Repository: anime_character.NewAnimeCharacterRepository(db),
		Options:    opt,
		Producer:   prod,
	}
}

func (p *PulsarAnimeCharacterPostgresProcessor) Process(ctx context.Context, data Payload) error {
	if data.Before == nil && data.After != nil {
		newChar, err := p.parseToEntity(ctx, *data.After)
		if err != nil {
			return err
		}
		if err := p.Repository.Upsert(newChar); err != nil {
			return err
		}

		jsonChar, err := json.Marshal(ProducerPayload{
			Action: CreateAction,
			Data:   data.After,
		})
		if err != nil {
			return err
		}
		if err := p.Producer.Send(ctx, jsonChar); err != nil {
			return err
		}
	}

	if data.After == nil && data.Before != nil {
		oldChar, err := p.parseToEntity(ctx, *data.Before)
		if err != nil {
			return err
		}
		if err := p.Repository.Delete(oldChar); err != nil {
			if p.Options.NoErrorOnDelete {
				log.Println("WARN: error deleting from db:", err)
				return nil
			}
			return err
		}
		return nil
	}

	if data.Before != nil && data.After != nil {
		newChar, err := p.parseToEntity(ctx, *data.After)
		if err != nil {
			return err
		}
		if err := p.Repository.Upsert(newChar); err != nil {
			return err
		}
	}

	if data.Before != nil && data.After == nil {
		log.Println("WARN: data.After is nil, skipping update")
	}

	return nil
}

func (p *PulsarAnimeCharacterPostgresProcessor) parseToEntity(ctx context.Context, data Schema) (*anime_character.AnimeCharacter, error) {
	return &anime_character.AnimeCharacter{
		ID:            data.Id,
		AnimeID:       ptrToString(data.AnimeID),
		Name:          ptrToString(data.Name),
		Role:          ptrToString(data.Role),
		Birthday:      ptrToString(data.Birthday),
		Zodiac:        ptrToString(data.Zodiac),
		Gender:        ptrToString(data.Gender),
		Race:          ptrToString(data.Race),
		Height:        ptrToString(data.Height),
		Weight:        ptrToString(data.Weight),
		Title:         ptrToString(data.Title),
		MartialStatus: ptrToString(data.MartialStatus),
		Summary:       ptrToString(data.Summary),
		Image:         ptrToString(data.Image),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil
}

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
