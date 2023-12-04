package utils

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
}

func LoadConfig(path string) *ConfigDatabase {
	var cfg ConfigDatabase

	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}
