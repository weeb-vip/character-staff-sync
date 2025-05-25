package pulsar_anime_staff_postgres_processor

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/weeb-vip/character-staff-sync/internal/db"
	"github.com/weeb-vip/character-staff-sync/internal/db/repositories/anime_staff"
	"github.com/weeb-vip/character-staff-sync/internal/producer"
)

type Options struct {
	NoErrorOnDelete bool
}

type PulsarAnimeStaffPostgresProcessorImpl interface {
	Process(ctx context.Context, data Payload) error
	parseToEntity(ctx context.Context, data Schema) (*anime_staff.AnimeStaff, error)
}

type PulsarAnimeStaffPostgresProcessor struct {
	Repository anime_staff.AnimeStaffRepositoryImpl
	Options    Options
	Producer   producer.Producer[Schema]
}

func NewPulsarAnimeStaffPostgresProcessor(opt Options, db *db.DB, prod producer.Producer[Schema]) PulsarAnimeStaffPostgresProcessorImpl {
	return &PulsarAnimeStaffPostgresProcessor{
		Repository: anime_staff.NewAnimeStaffRepository(db),
		Options:    opt,
		Producer:   prod,
	}
}

func (p *PulsarAnimeStaffPostgresProcessor) Process(ctx context.Context, data Payload) error {
	if data.Before == nil && data.After != nil {
		newStaff, err := p.parseToEntity(ctx, *data.After)
		if err != nil {
			return err
		}
		if err := p.Repository.Upsert(newStaff); err != nil {
			return err
		}

		if err != nil {
			return err
		}
		image := ""
		if data.After.Image != nil {
			image = *data.After.Image
		}
		payload := producer.ImageSchema{
			Name: *data.After.GivenName + "_" + *data.After.FamilyName,
			URL:  image,
			Type: producer.DataTypeStaff,
		}

		payloadBytes, err := json.Marshal(payload)
		if data.After.Image != nil {
			if err := p.Producer.Send(ctx, payloadBytes); err != nil {
				return err
			}
		}
	}

	if data.After == nil && data.Before != nil {
		oldStaff, err := p.parseToEntity(ctx, *data.Before)
		if err != nil {
			return err
		}
		if err := p.Repository.Delete(oldStaff); err != nil {
			if p.Options.NoErrorOnDelete {
				log.Println("WARN: error deleting from db:", err)
				return nil
			}
			return err
		}
		return nil
	}

	if data.Before != nil && data.After != nil {
		newStaff, err := p.parseToEntity(ctx, *data.After)
		if err != nil {
			return err
		}
		if err := p.Repository.Upsert(newStaff); err != nil {
			return err
		}
	}

	if data.Before != nil && data.After == nil {
		log.Println("WARN: data.After is nil, skipping update")
	}

	return nil
}

func (p *PulsarAnimeStaffPostgresProcessor) parseToEntity(ctx context.Context, data Schema) (*anime_staff.AnimeStaff, error) {
	return &anime_staff.AnimeStaff{
		ID:         data.Id,
		GivenName:  ptrToString(data.GivenName),
		FamilyName: ptrToString(data.FamilyName),
		Image:      ptrToString(data.Image),
		Birthday:   ptrToString(data.Birthday),
		BirthPlace: ptrToString(data.BirthPlace),
		BloodType:  ptrToString(data.BloodType),
		Hobbies:    ptrToString(data.Hobbies),
		Summary:    ptrToString(data.Summary),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}
func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
