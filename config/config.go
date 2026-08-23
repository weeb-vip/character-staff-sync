package config

import (
	"github.com/jinzhu/configor"
)

type Config struct {
	AppConfig   AppConfig
	DBConfig    DBConfig
	KafkaConfig KafkaConfig
	NatsConfig  NatsConfig
	FFConfig    FFConfig
}

type AppConfig struct {
	APPName string `default:"anime-api"`
	Port    int    `env:"PORT" default:"3000"`
	Version string `default:"x.x.x"`
}

type DBConfig struct {
	Host     string `default:"localhost" env:"DBHOST"`
	DataBase string `default:"weeb" env:"DBNAME"`
	User     string `default:"weeb" env:"DBUSERNAME"`
	Password string `required:"true" env:"DBPASSWORD" default:"mysecretpassword"`
	Port     uint   `default:"5432" env:"DBPORT"`
	SSLMode  string `default:"require" env:"DBSSL"`
}

type KafkaConfig struct {
	ConsumerGroupName string `default:"image-sync-group" env:"KAFKA_CONSUMER_GROUP_NAME"`
	BootstrapServers  string `default:"localhost:9092" env:"KAFKA_BOOTSTRAP_SERVERS"`
	// Where a consumer starts when it has no valid committed offset -- including
	// when retention has deleted past the one it had.
	//
	// earliest, not the rdkafka default of latest. This field did not exist and
	// the handlers passed nil, so a consumer whose offset fell behind the low
	// watermark jumped to the end of the topic and silently skipped everything in
	// between. It kept reporting lag and consuming nothing, which reads as a stall
	// rather than a reset. anime-sync has always set this; this service did not.
	Offset        string `default:"earliest" env:"KAFKA_OFFSET"`
	Topic         string `default:"anime-db.public.anime_staff" env:"KAFKA_TOPIC"`
	ProducerTopic string `default:"image-sync" env:"KAFKA_PRODUCER_TOPIC"`
}

type FFConfig struct {
	APIKey  string `default:"" env:"FF_API_KEY"`
	BaseURL string `default:"http://flagsmith-api.weeb.svc.cluster.local" env:"FF_BASE_URL"`
}

// NatsConfig mirrors KafkaConfig, so moving a service between the two is one
// obvious substitution per setting. A NATS subject is what Kafka calls a topic,
// and Debezium publishes to subjects named exactly like the topics it wrote, so
// the values carry over unchanged.
type NatsConfig struct {
	URL string `default:"nats://localhost:4222" env:"NATSURL"`

	// The durable consumer name -- the closest equivalent to a Kafka consumer
	// group. Left empty the consumer is ephemeral and loses its position on
	// restart.
	ConsumerGroupName string `default:"character-staff-sync-nats" env:"NATSCONSUMERGROUPNAME"`

	// Bind to Debezium's stream rather than deriving one from the subject.
	// JetStream refuses two streams whose subjects overlap, so creating a
	// per-subject stream would collide with the one Debezium declares over
	// anime-db-staging.>.
	StreamName string `default:"ANIMEDBSTAGING" env:"NATSSTREAMNAME"`

	Offset string `default:"earliest" env:"NATSOFFSET"`

	Subject string `default:"anime-db-staging.public.anime_character" env:"NATSSUBJECT"`

	// Outbound. Not CDC, so nothing else declares a stream over it and the
	// driver creates one as needed.
	ProducerSubject string `default:"image-sync" env:"NATSPRODUCERSUBJECT"`
}

func LoadConfigOrPanic() Config {
	var config = Config{}
	configor.Load(&config, "config/config.dev.json")

	return config
}
