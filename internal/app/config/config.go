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

type flagConfig struct {
	ServerAddress   *string
	BaseURL         *string
	FileStoragePath *string
}

var f flagConfig

func init() {
	f.ServerAddress = flag.String("a", "localhost:8080", "server address")
	f.BaseURL = flag.String("b", "sh", "base url")
	f.FileStoragePath = flag.String("f", "internal/app/storage/storage.txt", "file storage path")
}

func ParseConfig() (Config, error) {
	flag.Parse()
	Conf.ServerAddress = *f.ServerAddress
	Conf.FileStoragePath = *f.FileStoragePath
	Conf.BaseURL = *f.BaseURL

	err := env.Parse(&Conf)
	if err != nil {
		return Config{}, err
	}

	if Conf.ServerAddress == "" {
		Conf.ServerAddress = "localhost:8080"
	}

	if Conf.ServerAddress != "" {
		Conf.ServerAddress += "/"
	}

	if Conf.BaseURL != "" {
		Conf.BaseURL += "/"
	}

	return Conf, nil
}
