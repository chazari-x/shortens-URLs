package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

var Conf Config

type Flag struct {
	ServerAddress   *string
	BaseURL         *string
	FileStoragePath *string
}

var F Flag

func init() {
	F.ServerAddress = flag.String("a", "localhost:8080", "server address")
	F.BaseURL = flag.String("b", "sh", "base url")
	F.FileStoragePath = flag.String("f", "internal/app/storage/storage.txt", "file storage path")
}

func ParseConfig() (Config, error) {
	flag.Parse()
	Conf.ServerAddress = *F.ServerAddress
	Conf.FileStoragePath = *F.FileStoragePath
	Conf.BaseURL = *F.BaseURL

	err := env.Parse(&Conf)
	if err != nil {
		return Config{}, err
	}

	if Conf.ServerAddress == "" {
		Conf.ServerAddress = "localhost:8080"
	}

	return Conf, nil
}
