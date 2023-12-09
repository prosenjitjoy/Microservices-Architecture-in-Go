package util

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type ConfigDatabase struct {
	PulsarURL         string        `env:"PULSAR_URL" env-required:"true"`
	TopicName         string        `env:"TOPIC_NAME" env-required:"true"`
	SubscriberName    string        `env:"SUBSCRIBER_NAME" env-required:"true"`
	ConnectionTimeout time.Duration `env:"CONNECTION_TIMEOUT" env-required:"true"`
	OperationTimeout  time.Duration `env:"OPERATION_TIMEOUT" env-required:"true"`
	DatabaseURL       string        `env:"DATABASE_URL" env-required:"true"`
	MetadataPort      int           `env:"METADATA_PORT" env-required:"true"`
	RatingPort        int           `env:"RATING_PORT" env-required:"true"`
	MoviePort         int           `env:"MOVIE_PORT" env-required:"true"`
	Host              string        `env:"HOST" env-required:"true"`
	ConsulURL         string        `env:"CONSUL_URL" env-required:"true"`
	JaegerURL         string        `env:"JAEGER_URL" env-required:"true"`
	Environment       string        `env:"ENVIRONMENT" env-required:"true"`
}

func LoadConfig(path string) *ConfigDatabase {
	var cfg ConfigDatabase

	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}
