package utils

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type ConfigDatabase struct {
	PulsarURL         string        `yaml:"pulsar_url" env:"PULSAR_URL"`
	TopicName         string        `yaml:"topic_name" env:"TOPIC_NAME"`
	SubscriberName    string        `yaml:"subscriber_name" env:"SUBSCRIBER_NAME"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout" env:"CONNECTION_TIMEOUT"`
	OperationTimeout  time.Duration `yaml:"operation_timeout" env:"OPERATION_TIMEOUT"`
}

func LoadConfig() *ConfigDatabase {
	var cfg ConfigDatabase

	err := cleanenv.ReadConfig("config.yaml", &cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}
