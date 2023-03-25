package config

import (
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

func GetConfig() Config {
	var c Config

	err := env.Parse(&c)
	if err != nil {
		log.Fatal(err)
	}

	return c
}
