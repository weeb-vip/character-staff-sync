package eventing

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/ThatCatDev/ep/v2/drivers"
	epKafka "github.com/ThatCatDev/ep/v2/drivers/kafka"
	"github.com/ThatCatDev/ep/v2/event"
	"github.com/ThatCatDev/ep/v2/middleware"
	"github.com/ThatCatDev/ep/v2/middlewares/kafka/backoffretry"
	"github.com/ThatCatDev/ep/v2/processor"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/weeb-vip/character-staff-sync/config"
	"github.com/weeb-vip/character-staff-sync/internal/db"
	"github.com/weeb-vip/character-staff-sync/internal/db/repositories/anime_staff"
	"github.com/weeb-vip/character-staff-sync/internal/logger"
	"github.com/weeb-vip/character-staff-sync/internal/services/staff_processor"
	"go.uber.org/zap"
)

func EventingAnimeStaffKafka() error {
	cfg := config.LoadConfigOrPanic()
	ctx := context.Background()
	log := logger.Get()
	ctx = logger.WithCtx(ctx, log)

	kafkaConfig := &epKafka.KafkaConfig{
		ConsumerGroupName:        cfg.KafkaConfig.ConsumerGroupName,
		BootstrapServers:         cfg.KafkaConfig.BootstrapServers,
		SaslMechanism:            nil,
		SecurityProtocol:         nil,
		Username:                 nil,
		Password:                 nil,
		ConsumerSessionTimeoutMs: nil,
		ConsumerAutoOffsetReset:  nil,
		ClientID:                 nil,
		Debug:                    nil,
	}

	driver := epKafka.NewKafkaDriver(kafkaConfig)
	defer func(driver drivers.Driver[*kafka.Message]) {
		err := driver.Close()
		if err != nil {
			log.Error("Error closing Kafka driver", zap.String("error", err.Error()))
		} else {
			log.Info("Kafka driver closed successfully")
		}
	}(driver)

	database := db.NewDB(cfg.DBConfig)

	animeStaffRepo := anime_staff.NewAnimeStaffRepository(database)

	posgresProcessorOptions := staff_processor.Options{
		NoErrorOnDelete: true,
	}

	staffProcessor := staff_processor.NewStaffProcessor(posgresProcessorOptions, animeStaffRepo, KafkaProducer(ctx, driver, cfg.KafkaConfig.ProducerTopic))

	processorInstance := processor.NewProcessor[*kafka.Message, staff_processor.Payload](driver, cfg.KafkaConfig.Topic, staffProcessor.Process)

	log.Info("initializing backoff retry middleware", zap.String("topic", cfg.KafkaConfig.Topic))
	backoffRetryInstance := backoffretry.NewBackoffRetry[staff_processor.Payload](driver, backoffretry.Config{
		MaxRetries: 3,
		HeaderKey:  "retry",
		RetryQueue: cfg.KafkaConfig.Topic + "-retry",
	})

	log.Info("Starting Kafka processor", zap.String("topic", cfg.KafkaConfig.Topic))
	err := processorInstance.
		AddMiddleware(NewLoggerMiddleware[*kafka.Message, staff_processor.Payload]().Process).
		AddMiddleware(NewTransformMiddleware[*kafka.Message, staff_processor.Payload]().Process).
		AddMiddleware(backoffRetryInstance.Process).
		Run(ctx)

	if err != nil && ctx.Err() == nil { // Ignore error if caused by context cancellation
		log.Error("Error consuming messages", zap.String("error", err.Error()))
		return err
	}

	return nil
}

type LoggerMiddleware[DM any, M any] struct{}

func NewLoggerMiddleware[DM any, M any]() *LoggerMiddleware[DM, M] {
	return &LoggerMiddleware[DM, M]{}
}

func (f *LoggerMiddleware[DM, M]) Process(ctx context.Context, data event.Event[*kafka.Message, M], next middleware.Handler[*kafka.Message, M]) (*event.Event[*kafka.Message, M], error) {
	// if error log it
	log := logger.FromCtx(ctx)

	result, err := next(ctx, data)
	if err != nil {
		log.Error("Error processing message", zap.Error(err))
	} else {
		log.Info("Message processed successfully")
	}

	jsonPayload, err := json.Marshal(result.Payload)
	log.Info("Processing message", zap.String("value", string(jsonPayload)))
	if err != nil {
		log.Error("Error processing message", zap.String("value", string(jsonPayload)), zap.Error(err))
	} else {
		log.Info("Successfully processed message", zap.String("value", string(jsonPayload)))
	}
	return result, err
}

type TransformMiddleware[DM any, M any] struct {
}

func NewTransformMiddleware[DM any, M any]() *TransformMiddleware[DM, M] {
	return &TransformMiddleware[DM, M]{}
}

func (f *TransformMiddleware[DM, M]) Process(ctx context.Context, data event.Event[*kafka.Message, M], next middleware.Handler[*kafka.Message, M]) (*event.Event[*kafka.Message, M], error) {
	log := logger.FromCtx(ctx)
	log.Info("starting TransformMiddleware")

	if valueRaw, exists := data.RawData["Value"]; exists {
		if valueStr, ok := valueRaw.(string); ok {
			log.Info("Value key found in RawData", zap.Any("value", valueRaw))
			decodedBytes, err := base64.StdEncoding.DecodeString(valueStr)
			if err != nil {
				log.Error("Failed to decode base64 value", zap.Error(err))
				return nil, err
			}

			log.Info("Decoding base64 value", zap.String("decodedBytes", string(decodedBytes)))

			var debeziumMessage struct {
				Schema  interface{} `json:"schema"`
				Payload M           `json:"payload"`
			}
			if err := json.Unmarshal(decodedBytes, &debeziumMessage); err != nil {
				log.Error("Failed to unmarshal decoded payload", zap.Error(err))
				return nil, err
			}
			if payload, ok := interface{}(debeziumMessage.Payload).(M); ok {
				data.Payload = payload
			} else {
				log.Error("Failed to cast payload to expected type")
				return nil, err
			}

			log.Info("Successfully decoded base64 value and updated payload", zap.Any("payload", data.Payload))
		} else {
			log.Warn("Value in RawData is not a string")
		}
	} else {
		log.Warn("Value key not found in RawData")
	}

	return next(ctx, data)
}
