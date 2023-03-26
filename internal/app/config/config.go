package config

import (
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

var Conf Config

func ParseConfig() Config {
	err := env.Parse(&Conf)
	if err != nil {
		log.Fatal(err)
	}

	if Conf.ServerAddress == "" {
		Conf.ServerAddress = "localhost:8080"
	}

	return Conf
}
