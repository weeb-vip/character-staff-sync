package pulsar_anime_character_postgres_processor

import (
	"context"
	"encoding/json"
	"github.com/Flagsmith/flagsmith-go-client/v2"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/weeb-vip/character-staff-sync/internal"
	"github.com/weeb-vip/character-staff-sync/internal/logger"
	"go.uber.org/zap"
	"time"

	"github.com/weeb-vip/character-staff-sync/internal/db"
	"github.com/weeb-vip/character-staff-sync/internal/db/repositories/anime_character"
	"github.com/weeb-vip/character-staff-sync/internal/producer"
)

type Options struct {
	NoErrorOnDelete bool
}

type PulsarAnimeCharacterPostgresProcessor interface {
	Process(ctx context.Context, data Payload) error
	parseToEntity(ctx context.Context, data Schema) (*anime_character.AnimeCharacter, error)
}

type PulsarAnimeCharacterPostgresProcessorImpl struct {
	Repository    anime_character.AnimeCharacterRepository
	Options       Options
	Producer      producer.Producer[Schema]
	KafkaProducer func(ctx context.Context, message *kafka.Message) error
}

func NewPulsarAnimeCharacterPostgresProcessor(opt Options, db *db.DB, prod producer.Producer[Schema], kafkaProducer func(ctx context.Context, message *kafka.Message) error) PulsarAnimeCharacterPostgresProcessor {
	return &PulsarAnimeCharacterPostgresProcessorImpl{
		Repository:    anime_character.NewAnimeCharacterRepository(db),
		Options:       opt,
		Producer:      prod,
		KafkaProducer: kafkaProducer,
	}
}

func (p *PulsarAnimeCharacterPostgresProcessorImpl) Process(ctx context.Context, data Payload) error {
	log := logger.FromCtx(ctx)

	log.Info("Gettting flagsmith client from context")
	flagsmithClient, _ := ctx.Value(internal.FFClient{}).(*flagsmith.Client)

	log.Info("Getting environment flags from flagsmith client")
	flags, _ := flagsmithClient.GetEnvironmentFlags()

	log.Info("Checking if feature 'enable_kafka' is enabled")
	isEnabled, _ := flags.IsFeatureEnabled("enable_kafka")
	log.Info("Feature 'enable_kafka' is enabled", zap.Bool("isEnabled", isEnabled))

	if data.Before == nil && data.After != nil {
		newChar, err := p.parseToEntity(ctx, *data.After)
		if err != nil {
			return err
		}
		if err := p.Repository.Upsert(newChar); err != nil {
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
			Name: *data.After.Name + "_" + *data.After.AnimeID,
			URL:  image,
			Type: producer.DataTypeCharacter,
		}
		
		payloadBytes, _ := json.Marshal(payload)
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
		oldChar, err := p.parseToEntity(ctx, *data.Before)
		if err != nil {
			return err
		}
		if err := p.Repository.Delete(oldChar); err != nil {
			if p.Options.NoErrorOnDelete {
				log.Warn("WARN: error deleting from db:", zap.Error(err))
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
		log.Warn("WARN: data.After is nil, skipping update")
	}

	return nil
}

func (p *PulsarAnimeCharacterPostgresProcessorImpl) parseToEntity(ctx context.Context, data Schema) (*anime_character.AnimeCharacter, error) {
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
