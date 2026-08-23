package eventing

import (
	"context"

	"github.com/ThatCatDev/ep/v2/drivers"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/weeb-vip/character-staff-sync/internal/logger"
	"go.uber.org/zap"
)

// KafkaProducer returns a Produce closure bound to one topic.
//
// It lived in handler_anime_character.go, the Pulsar handler, even though all
// three Kafka handlers depend on it -- so deleting the Pulsar path took the
// Kafka path's producer with it. It has a file of its own now.
//
// It takes the encoded value rather than a *kafka.Message because the
// processors are generic over the driver message: building the transport's
// message is this function's job, not theirs.
func KafkaProducer(ctx context.Context, driver drivers.Driver[*kafka.Message], topic string) func(ctx context.Context, value []byte) error {
	return func(ctx context.Context, value []byte) error {
		log := logger.FromCtx(ctx)
		log.Info("Producing message to Kafka", zap.String("topic", topic), zap.String("value", string(value)))
		if err := driver.Produce(ctx, topic, &kafka.Message{Value: value}); err != nil {
			log.Error("Failed to produce message", zap.String("topic", topic), zap.Error(err))

			return err
		}

		return nil
	}
}
