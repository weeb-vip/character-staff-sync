package character_staff_link_processor

import (
	"context"
	"encoding/json"
	"github.com/ThatCatDev/ep/v2/event"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/weeb-vip/character-staff-sync/internal/db/repositories/anime_character_staff_link"
	"github.com/weeb-vip/character-staff-sync/internal/logger"
	"go.uber.org/zap"
	"time"
)

type Options struct {
	NoErrorOnDelete bool
}

type CharacterStaffLinkProcessorImpl interface {
	Process(ctx context.Context, data event.Event[*kafka.Message, Payload]) (event.Event[*kafka.Message, Payload], error)
	parseToEntity(ctx context.Context, data Schema) (*anime_character_staff_link.AnimeCharacterStaffLink, error)
}

type CharacterStaffLinkProcessor struct {
	Repository    anime_character_staff_link.AnimeCharacterStaffLinkRepositoryImpl
	Options       Options
	KafkaProducer func(ctx context.Context, message *kafka.Message) error
}

func NewCharacterStaffLinkProcessor(opt Options, repo anime_character_staff_link.AnimeCharacterStaffLinkRepositoryImpl, kafkaProducer func(ctx context.Context, message *kafka.Message) error) CharacterStaffLinkProcessorImpl {
	return &CharacterStaffLinkProcessor{
		Repository:    repo,
		Options:       opt,
		KafkaProducer: kafkaProducer,
	}
}

func (p *CharacterStaffLinkProcessor) Process(ctx context.Context, data event.Event[*kafka.Message, Payload]) (event.Event[*kafka.Message, Payload], error) {
	log := logger.FromCtx(ctx)

	payload := data.Payload

	if payload.Before == nil && payload.After != nil {
		newLink, err := p.parseToEntity(ctx, *payload.After)
		if err != nil {
			return data, err
		}
		if err := p.Repository.Upsert(newLink); err != nil {
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
		oldLink, err := p.parseToEntity(ctx, *payload.Before)
		if err != nil {
			return data, err
		}
		if err := p.Repository.Delete(oldLink); err != nil {
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
		newLink, err := p.parseToEntity(ctx, *payload.After)
		if err != nil {
			return data, err
		}
		if err := p.Repository.Upsert(newLink); err != nil {
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

func (p *CharacterStaffLinkProcessor) parseToEntity(ctx context.Context, data Schema) (*anime_character_staff_link.AnimeCharacterStaffLink, error) {
	return &anime_character_staff_link.AnimeCharacterStaffLink{
		ID:              data.ID,
		CharacterID:     data.CharacterID,
		StaffID:         data.StaffID,
		CharacterName:   ptrToString(data.CharacterName),
		StaffGivenName:  ptrToString(data.StaffGivenName),
		StaffFamilyName: ptrToString(data.StaffFamilyName),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}, nil
}

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}