package eventing

import (
	"context"
	"github.com/ThatCatDev/ep/v2/drivers"
	epKafka "github.com/ThatCatDev/ep/v2/drivers/kafka"
	"github.com/ThatCatDev/ep/v2/middlewares/kafka/backoffretry"
	"github.com/ThatCatDev/ep/v2/processor"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/weeb-vip/character-staff-sync/config"
	"github.com/weeb-vip/character-staff-sync/internal/db"
	"github.com/weeb-vip/character-staff-sync/internal/db/repositories/anime_character_staff_link"
	"github.com/weeb-vip/character-staff-sync/internal/logger"
	"github.com/weeb-vip/character-staff-sync/internal/services/character_staff_link_processor"
	"go.uber.org/zap"
)

func EventingAnimeCharacterStaffLinkKafka() error {
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

	characterStaffLinkRepo := anime_character_staff_link.NewAnimeCharacterStaffLinkRepository(database)

	processorOptions := character_staff_link_processor.Options{
		NoErrorOnDelete: true,
	}

	linkProcessor := character_staff_link_processor.NewCharacterStaffLinkProcessor(processorOptions, characterStaffLinkRepo, KafkaProducer(ctx, driver, cfg.KafkaConfig.ProducerTopic))

	processorInstance := processor.NewProcessor[*kafka.Message, character_staff_link_processor.Payload](driver, cfg.KafkaConfig.Topic, linkProcessor.Process)

	log.Info("initializing backoff retry middleware", zap.String("topic", cfg.KafkaConfig.Topic))
	backoffRetryInstance := backoffretry.NewBackoffRetry[character_staff_link_processor.Payload](driver, backoffretry.Config{
		MaxRetries: 3,
		HeaderKey:  "retry",
		RetryQueue: cfg.KafkaConfig.Topic + "-retry",
	})

	log.Info("Starting Kafka processor", zap.String("topic", cfg.KafkaConfig.Topic))
	err := processorInstance.
		AddMiddleware(NewLoggerMiddleware[*kafka.Message, character_staff_link_processor.Payload]().Process).
		AddMiddleware(NewTransformMiddleware[*kafka.Message, character_staff_link_processor.Payload]().Process).
		AddMiddleware(backoffRetryInstance.Process).
		Run(ctx)

	if err != nil && ctx.Err() == nil { // Ignore error if caused by context cancellation
		log.Error("Error consuming messages", zap.String("error", err.Error()))
		return err
	}

	return nil
}
