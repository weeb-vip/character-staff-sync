package staff_processor

import (
	"context"
	"encoding/json"
	"github.com/ThatCatDev/ep/v2/event"
	"github.com/weeb-vip/character-staff-sync/internal/db/repositories/anime_staff"
	"github.com/weeb-vip/character-staff-sync/internal/logger"
	"github.com/weeb-vip/character-staff-sync/internal/producer"
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
type StaffProcessor[DM any] interface {
	Process(ctx context.Context, data event.Event[DM, Payload]) (event.Event[DM, Payload], error)
}

type StaffProcessorImpl[DM any] struct {
	Repository anime_staff.AnimeStaffRepository
	Options    Options
	Producer   func(ctx context.Context, value []byte) error
}

func NewStaffProcessor[DM any](opt Options, repo anime_staff.AnimeStaffRepository, producer func(ctx context.Context, value []byte) error) StaffProcessor[DM] {
	return &StaffProcessorImpl[DM]{
		Repository: repo,
		Options:    opt,
		Producer:   producer,
	}
}

func (p *StaffProcessorImpl[DM]) Process(ctx context.Context, data event.Event[DM, Payload]) (event.Event[DM, Payload], error) {
	log := logger.FromCtx(ctx)

	payload := data.Payload

	if payload.Before == nil && payload.After != nil {
		newStaff, err := p.parseToEntity(ctx, *payload.After)
		if err != nil {
			return data, err
		}
		log.Info("INFO: newStaff", zap.String("ID", newStaff.ID),
			zap.String("GivenName", newStaff.GivenName),
			zap.String("FamilyName", newStaff.FamilyName))
		if err := p.Repository.Upsert(newStaff); err != nil {
			return data, err
		}

		if err != nil {
			return data, err
		}
		image := ""
		if payload.After.Image != nil {
			image = *payload.After.Image
		}
		imagePayload := producer.ImagePayload{
			Data: producer.ImageSchema{
				ID:   newStaff.ID,
				Name: *payload.After.GivenName + "_" + *payload.After.FamilyName,
				URL:  image,
				Type: producer.DataTypeStaff,
			},
		}

		payloadBytes, err := json.Marshal(imagePayload)
		if payload.After.Image != nil {
			log.Info("Sending update to producer", zap.String("title", imagePayload.Data.Name), zap.String("imageURL", imagePayload.Data.URL))

			err = p.Producer(ctx, payloadBytes)

			if err != nil {
				log.Error("Error sending message to producer", zap.Error(err))
				return data, err
			}
		}
	}

	if payload.After == nil && payload.Before != nil {
		oldStaff, err := p.parseToEntity(ctx, *payload.Before)
		if err != nil {
			return data, err
		}
		if err := p.Repository.Delete(oldStaff); err != nil {
			if p.Options.NoErrorOnDelete {
				log.Warn("WARN: error deleting from db:", zap.Error(err))
				return data, nil
			}
			return data, err
		}
		return data, nil
	}

	if payload.Before != nil && payload.After != nil {
		newStaff, err := p.parseToEntity(ctx, *payload.After)
		if err != nil {
			return data, err
		}
		// log newStaff
		log.Info("INFO: newStaff", zap.String("ID", newStaff.ID),
			zap.String("GivenName", newStaff.GivenName),
			zap.String("FamilyName", newStaff.FamilyName))
		if err := p.Repository.Upsert(newStaff); err != nil {
			return data, err
		}
	}

	if payload.Before != nil && payload.After == nil {
		log.Warn("WARN: payload.After is nil, skipping update")
	}

	return data, nil
}

func (p *StaffProcessorImpl[DM]) parseToEntity(ctx context.Context, data Schema) (*anime_staff.AnimeStaff, error) {
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
