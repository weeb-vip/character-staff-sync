package pulsar_anime_staff_postgres_processor

import (
	"context"
	"encoding/json"
	"github.com/Flagsmith/flagsmith-go-client/v2"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/weeb-vip/character-staff-sync/internal"
	"github.com/weeb-vip/character-staff-sync/internal/db"
	"github.com/weeb-vip/character-staff-sync/internal/db/repositories/anime_staff"
	"github.com/weeb-vip/character-staff-sync/internal/logger"
	"github.com/weeb-vip/character-staff-sync/internal/producer"
	"go.uber.org/zap"
	"time"
)

type Options struct {
	NoErrorOnDelete bool
}

type PulsarAnimeStaffPostgresProcessor interface {
	Process(ctx context.Context, data Payload) error
	parseToEntity(ctx context.Context, data Schema) (*anime_staff.AnimeStaff, error)
}

type PulsarAnimeStaffPostgresProcessorImpl struct {
	Repository    anime_staff.AnimeStaffRepository
	Options       Options
	Producer      producer.Producer[Schema]
	KafkaProducer func(ctx context.Context, message *kafka.Message) error
}

func NewPulsarAnimeStaffPostgresProcessor(opt Options, db *db.DB, prod producer.Producer[Schema], kafkaProducer func(ctx context.Context, message *kafka.Message) error) PulsarAnimeStaffPostgresProcessor {
	return &PulsarAnimeStaffPostgresProcessorImpl{
		Repository:    anime_staff.NewAnimeStaffRepository(db),
		Options:       opt,
		Producer:      prod,
		KafkaProducer: kafkaProducer,
	}
}

func (p *PulsarAnimeStaffPostgresProcessorImpl) Process(ctx context.Context, data Payload) error {
	log := logger.FromCtx(ctx)

	log.Info("Gettting flagsmith client from context")
	flagsmithClient, _ := ctx.Value(internal.FFClient{}).(*flagsmith.Client)

	log.Info("Getting environment flags from flagsmith client")
	flags, _ := flagsmithClient.GetEnvironmentFlags()

	log.Info("Checking if feature 'enable_kafka' is enabled")
	isEnabled, _ := flags.IsFeatureEnabled("enable_kafka")
	log.Info("Feature 'enable_kafka' is enabled", zap.Bool("isEnabled", isEnabled))

	if data.Before == nil && data.After != nil {
		newStaff, err := p.parseToEntity(ctx, *data.After)
		if err != nil {
			return err
		}
		log.Info("INFO: newStaff", zap.String("ID", newStaff.ID),
			zap.String("GivenName", newStaff.GivenName),
			zap.String("FamilyName", newStaff.FamilyName))
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
			log.Info("Sending update to producer", zap.String("title", payload.Name), zap.String("imageURL", payload.URL))

			if isEnabled {
				err = p.KafkaProducer(ctx, &kafka.Message{
					Value: payloadBytes,
				})
			} else {
				err = p.Producer.Send(ctx, payloadBytes)
			}
			if err != nil {
				log.Error("Error sending message to producer", zap.Error(err))
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
				log.Warn("WARN: error deleting from db:", zap.Error(err))
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
		// log newStaff
		log.Info("INFO: newStaff", zap.String("ID", newStaff.ID),
			zap.String("GivenName", newStaff.GivenName),
			zap.String("FamilyName", newStaff.FamilyName))
		if err := p.Repository.Upsert(newStaff); err != nil {
			return err
		}
	}

	if data.Before != nil && data.After == nil {
		log.Warn("WARN: data.After is nil, skipping update")
	}

	return nil
}

func (p *PulsarAnimeStaffPostgresProcessorImpl) parseToEntity(ctx context.Context, data Schema) (*anime_staff.AnimeStaff, error) {
	return &anime_staff.AnimeStaff{
		ID:         data.Id,
		Language:   ptrToString(data.Language),
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
