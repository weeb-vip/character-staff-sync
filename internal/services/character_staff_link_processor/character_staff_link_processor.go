package character_staff_link_processor

import (
	"context"
	"encoding/json"
	"github.com/ThatCatDev/ep/v2/event"
	"github.com/weeb-vip/character-staff-sync/internal/db"
	"github.com/weeb-vip/character-staff-sync/internal/db/repositories/anime_character_staff_link"
	"github.com/weeb-vip/character-staff-sync/internal/logger"
	"go.uber.org/zap"
	"time"
)

type Options struct {
	NoErrorOnDelete bool
}

// The driver message type is a parameter because the processor never looks at
// it. Nothing here reads DriverMessage, RawData or Headers -- only Payload,
// which the transform middleware has already filled in. Hard-coding
// *kafka.Message meant this could not be reused over NATS despite none of the
// logic being Kafka-specific.
//
// Producers take the encoded value rather than a driver message for the same
// reason: every call site only ever set Value, so building the transport's
// message belongs in the handler that knows which transport it is.
type CharacterStaffLinkProcessor[DM any] interface {
	Process(ctx context.Context, data event.Event[DM, Payload]) (event.Event[DM, Payload], error)
}

type CharacterStaffLinkProcessorImpl[DM any] struct {
	Repository    anime_character_staff_link.AnimeCharacterStaffLinkRepository
	Options       Options
	KafkaProducer func(ctx context.Context, value []byte) error
}

func NewCharacterStaffLinkProcessor[DM any](opt Options, repo anime_character_staff_link.AnimeCharacterStaffLinkRepository, kafkaProducer func(ctx context.Context, value []byte) error) CharacterStaffLinkProcessor[DM] {
	return &CharacterStaffLinkProcessorImpl[DM]{
		Repository:    repo,
		Options:       opt,
		KafkaProducer: kafkaProducer,
	}
}

func (p *CharacterStaffLinkProcessorImpl[DM]) Process(ctx context.Context, data event.Event[DM, Payload]) (event.Event[DM, Payload], error) {
	log := logger.FromCtx(ctx)

	payload := data.Payload

	if payload.Before == nil && payload.After != nil {
		newLink, err := p.parseToEntity(ctx, *payload.After)
		if err != nil {
			return data, err
		}
		if err := p.Repository.Upsert(newLink); err != nil {
			// The character or staff row this link points at has not arrived yet.
			// Debezium gives no ordering guarantee across tables, so this happens
			// occasionally and cannot be fixed by retrying: the parent arrives on
			// its own topic. Drop the event rather than block the consumer -- the
			// link comes back with the next update or snapshot.
			if db.IsForeignKeyViolation(err) {
				log.Warn("skipping link whose character or staff is not present",
					zap.String("character_id", newLink.CharacterID),
					zap.String("staff_id", newLink.StaffID),
					zap.String("link_id", newLink.ID))
				return data, nil
			}
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
			err = p.KafkaProducer(ctx, payloadBytes)
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
			err = p.KafkaProducer(ctx, payloadBytes)
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
			// The character or staff row this link points at has not arrived yet.
			// Debezium gives no ordering guarantee across tables, so this happens
			// occasionally and cannot be fixed by retrying: the parent arrives on
			// its own topic. Drop the event rather than block the consumer -- the
			// link comes back with the next update or snapshot.
			if db.IsForeignKeyViolation(err) {
				log.Warn("skipping link whose character or staff is not present",
					zap.String("character_id", newLink.CharacterID),
					zap.String("staff_id", newLink.StaffID),
					zap.String("link_id", newLink.ID))
				return data, nil
			}
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
			err = p.KafkaProducer(ctx, payloadBytes)
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

func (p *CharacterStaffLinkProcessorImpl[DM]) parseToEntity(ctx context.Context, data Schema) (*anime_character_staff_link.AnimeCharacterStaffLink, error) {
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
