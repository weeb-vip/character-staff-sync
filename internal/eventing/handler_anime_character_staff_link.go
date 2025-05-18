package eventing

import (
	"context"
	"fmt"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/weeb-vip/character-staff-sync/config"
	"github.com/weeb-vip/character-staff-sync/internal/db"
	"github.com/weeb-vip/character-staff-sync/internal/logger"
	"github.com/weeb-vip/character-staff-sync/internal/producer"
	"github.com/weeb-vip/character-staff-sync/internal/services/consumer"
	"github.com/weeb-vip/character-staff-sync/internal/services/processor"
	"github.com/weeb-vip/character-staff-sync/internal/services/pulsar_anime_character_staff_link_postgres_processor"
)

func EventingAnimeCharacterStaffLink() error {
	cfg := config.LoadConfigOrPanic()
	ctx := context.Background()
	log := logger.Get()

	database := db.NewDB(cfg.DBConfig)

	processorOptions := pulsar_anime_character_staff_link_postgres_processor.Options{
		NoErrorOnDelete: true,
	}

	linkProducer := producer.NewProducer[pulsar_anime_character_staff_link_postgres_processor.ProducerPayload](ctx, cfg.PulsarConfig)

	linkProcessor := pulsar_anime_character_staff_link_postgres_processor.NewPulsarAnimeCharacterStaffLinkPostgresProcessor(
		processorOptions,
		database,
		linkProducer,
	)

	messageProcessor := processor.NewProcessor[pulsar_anime_character_staff_link_postgres_processor.Payload]()

	linkConsumer := consumer.NewConsumer[pulsar_anime_character_staff_link_postgres_processor.Payload](ctx, cfg.PulsarConfig)

	log.Info("Starting anime character-staff link eventing")
	err := linkConsumer.Receive(ctx, func(ctx context.Context, msg pulsar.Message) error {
		return messageProcessor.Process(ctx, string(msg.Payload()), linkProcessor.Process)
	})
	if err != nil {
		log.Error(fmt.Sprintf("Error receiving character-staff link message: %v", err))
		return err
	}

	return nil
}
