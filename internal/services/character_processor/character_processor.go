package character_processor

import (
	"context"
	"encoding/json"
	"github.com/ThatCatDev/ep/v2/event"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/weeb-vip/character-staff-sync/internal/db/repositories/anime_character"
	"github.com/weeb-vip/character-staff-sync/internal/logger"
	"go.uber.org/zap"
	"time"
)

type Options struct {
	NoErrorOnDelete bool
}

type CharacterProcessor interface {
	Process(ctx context.Context, data event.Event[*kafka.Message, Payload]) (event.Event[*kafka.Message, Payload], error)
}

type CharacterProcessorImpl struct {
	Repository    anime_character.AnimeCharacterRepository
	Options       Options
	KafkaProducer func(ctx context.Context, message *kafka.Message) error
}

func NewCharacterProcessor(opt Options, repo anime_character.AnimeCharacterRepository, kafkaProducer func(ctx context.Context, message *kafka.Message) error) CharacterProcessor {
	return &CharacterProcessorImpl{
		Repository:    repo,
		Options:       opt,
		KafkaProducer: kafkaProducer,
	}
}

func (p *CharacterProcessorImpl) Process(ctx context.Context, data event.Event[*kafka.Message, Payload]) (event.Event[*kafka.Message, Payload], error) {
	log := logger.FromCtx(ctx)

	payload := data.Payload

	if payload.Before == nil && payload.After != nil {
		newChar, err := p.parseToEntity(ctx, *payload.After)
		if err != nil {
			return data, err
		}
		if err := p.Repository.Upsert(newChar); err != nil {
			return data, err
		}

		producerPayload := ProducerPayload{
			Action: CreateAction,
			Data:   payload.After,
		}

		payloadBytes, err := json.Marshal(producerPayload)
		if err != nil {
			log.Error("Error marshaling producer payload", zap.Error(err))
			return data, err
		}

		if p.KafkaProducer != nil {
			err = p.KafkaProducer(ctx, &kafka.Message{
				Value: payloadBytes,
			})
			if err != nil {
				log.Error("Error sending message to Kafka producer", zap.Error(err))
				return data, err
			}
		}
	}

	if payload.After == nil && payload.Before != nil {
		oldChar, err := p.parseToEntity(ctx, *payload.Before)
		if err != nil {
			return data, err
		}
		if err := p.Repository.Delete(oldChar); err != nil {
			if p.Options.NoErrorOnDelete {
				log.Warn("WARN: error deleting from db:", zap.Error(err))
				return data, nil
			}
			return data, err
		}

		producerPayload := ProducerPayload{
			Action: DeleteAction,
			Data:   payload.Before,
		}

		payloadBytes, err := json.Marshal(producerPayload)
		if err != nil {
			log.Error("Error marshaling producer payload", zap.Error(err))
			return data, err
		}

		if p.KafkaProducer != nil {
			err = p.KafkaProducer(ctx, &kafka.Message{
				Value: payloadBytes,
			})
			if err != nil {
				log.Error("Error sending message to Kafka producer", zap.Error(err))
				return data, err
			}
		}
		return data, nil
	}

	if payload.Before != nil && payload.After != nil {
		newChar, err := p.parseToEntity(ctx, *payload.After)
		if err != nil {
			return data, err
		}
		if err := p.Repository.Upsert(newChar); err != nil {
			return data, err
		}

		producerPayload := ProducerPayload{
			Action: UpdateAction,
			Data:   payload.After,
		}

		payloadBytes, err := json.Marshal(producerPayload)
		if err != nil {
			log.Error("Error marshaling producer payload", zap.Error(err))
			return data, err
		}

		if p.KafkaProducer != nil {
			err = p.KafkaProducer(ctx, &kafka.Message{
				Value: payloadBytes,
			})
			if err != nil {
				log.Error("Error sending message to Kafka producer", zap.Error(err))
				return data, err
			}
		}
	}

	if payload.Before != nil && payload.After == nil {
		log.Warn("WARN: payload.After is nil, skipping update")
	}

	return data, nil
}

func (p *CharacterProcessorImpl) parseToEntity(ctx context.Context, data Schema) (*anime_character.AnimeCharacter, error) {
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
