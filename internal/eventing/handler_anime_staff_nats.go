package eventing

import (
	"context"

	"github.com/ThatCatDev/ep/v2/drivers"
	epNats "github.com/ThatCatDev/ep/v2/drivers/nats"
	"github.com/ThatCatDev/ep/v2/middlewares/nats/backoffretry"
	"github.com/ThatCatDev/ep/v2/processor"
	"github.com/weeb-vip/character-staff-sync/config"
	"github.com/weeb-vip/character-staff-sync/internal/db"
	"github.com/weeb-vip/character-staff-sync/internal/db/repositories/anime_staff"
	"github.com/weeb-vip/character-staff-sync/internal/logger"
	"github.com/weeb-vip/character-staff-sync/internal/services/staff_processor"
	"go.uber.org/zap"
)

// EventingAnimeStaffNats is the NATS counterpart of EventingAnimeStaffKafka.
//
// A separate entry point rather than a flag, so production keeps running the
// Kafka command untouched while staging moves over. The two share the
// processor, the repository and the retry policy; only the transport differs.
func EventingAnimeStaffNats() error {
	cfg := config.LoadConfigOrPanic()
	ctx := context.Background()
	log := logger.Get()
	ctx = logger.WithCtx(ctx, log)

	natsConfig := &epNats.Config{
		URL:               cfg.NatsConfig.URL,
		ConsumerGroupName: cfg.NatsConfig.ConsumerGroupName,
		// Bind to Debezium's stream rather than deriving one per subject:
		// JetStream refuses two streams whose subjects overlap.
		StreamName:              cfg.NatsConfig.StreamName,
		ConsumerAutoOffsetReset: &cfg.NatsConfig.Offset,
	}

	driver := epNats.NewNatsDriver(natsConfig)
	defer func(driver drivers.Driver[*epNats.Message]) {
		if err := driver.Close(); err != nil {
			log.Error("Error closing NATS driver", zap.String("error", err.Error()))
		} else {
			log.Info("NATS driver closed successfully")
		}
	}(driver)

	// Publishing goes through its own driver so image-sync gets its own stream
	// instead of being forced into Debezium's. See NewNatsProducerDriver.
	producerDriver := NewNatsProducerDriver(cfg)
	defer func(d drivers.Driver[*epNats.Message]) {
		if err := d.Close(); err != nil {
			log.Error("Error closing NATS producer driver", zap.String("error", err.Error()))
		}
	}(producerDriver)

	database := db.NewDB(cfg.DBConfig)

	repo := anime_staff.NewAnimeStaffRepository(database)

	processorOptions := staff_processor.Options{
		NoErrorOnDelete: true,
	}

	procInstance := staff_processor.NewStaffProcessor[*epNats.Message](processorOptions, repo, NatsProducer(ctx, producerDriver, cfg.NatsConfig.ProducerSubject))

	processorInstance := processor.NewProcessor[*epNats.Message, staff_processor.Payload](driver, cfg.NatsConfig.Subject, procInstance.Process)

	log.Info("initializing backoff retry middleware", zap.String("subject", cfg.NatsConfig.Subject))
	backoffRetryInstance := backoffretry.NewBackoffRetry[staff_processor.Payload](driver, backoffretry.Config{
		MaxRetries: 3,
		HeaderKey:  "retry",
		RetryQueue: cfg.NatsConfig.Subject + "-retry",
	})

	log.Info("Starting NATS processor", zap.String("subject", cfg.NatsConfig.Subject))

	err := processorInstance.
		AddMiddleware(NewNatsLoggerMiddleware[staff_processor.Payload]().Process).
		AddMiddleware(NewNatsTransformMiddleware[staff_processor.Payload]().Process).
		AddMiddleware(backoffRetryInstance.Process).
		Run(ctx)

	if err != nil && ctx.Err() == nil { // Ignore error if caused by context cancellation
		log.Error("Error consuming messages", zap.String("error", err.Error()))

		return err
	}

	return nil
}
