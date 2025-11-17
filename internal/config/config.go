package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DBUrl       string
	RabbitMQUrl string
}

func Load() Config {
	viper.AutomaticEnv()

	return Config{
		DBUrl:       viper.GetString("DB_URL"),
		RabbitMQUrl: viper.GetString("RABBITMQ_URL"),
	}
}
